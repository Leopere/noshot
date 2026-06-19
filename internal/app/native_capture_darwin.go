//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c -fblocks
#cgo darwin LDFLAGS: -framework AppKit -framework CoreGraphics
#import <AppKit/AppKit.h>
#import <CoreGraphics/CoreGraphics.h>
#include <dlfcn.h>
#include <stdlib.h>

typedef CGImageRef (*NoShotDisplayCreateImageFn)(CGDirectDisplayID displayID);
typedef CGImageRef (*NoShotDisplayCreateImageForRectFn)(CGDirectDisplayID displayID, CGRect rect);
typedef CGImageRef (*NoShotWindowListCreateImageFn)(CGRect screenBounds, CGWindowListOption listOption, CGWindowID windowID, CGWindowImageOption imageOption);

static void *noshot_coregraphics(void) {
	static void *handle = NULL;
	if (handle == NULL) {
		handle = dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", RTLD_LAZY);
	}
	return handle;
}

static CGImageRef noshot_display_create_image(CGDirectDisplayID displayID) {
	void *handle = noshot_coregraphics();
	if (handle == NULL) {
		return NULL;
	}
	NoShotDisplayCreateImageFn fn = (NoShotDisplayCreateImageFn)dlsym(handle, "CGDisplayCreateImage");
	return fn == NULL ? NULL : fn(displayID);
}

static CGImageRef noshot_display_create_image_for_rect(CGDirectDisplayID displayID, CGRect rect) {
	void *handle = noshot_coregraphics();
	if (handle == NULL) {
		return NULL;
	}
	NoShotDisplayCreateImageForRectFn fn = (NoShotDisplayCreateImageForRectFn)dlsym(handle, "CGDisplayCreateImageForRect");
	return fn == NULL ? NULL : fn(displayID, rect);
}

static CGImageRef noshot_window_list_create_image(CGRect screenBounds, CGWindowListOption listOption, CGWindowID windowID, CGWindowImageOption imageOption) {
	void *handle = noshot_coregraphics();
	if (handle == NULL) {
		return NULL;
	}
	NoShotWindowListCreateImageFn fn = (NoShotWindowListCreateImageFn)dlsym(handle, "CGWindowListCreateImage");
	return fn == NULL ? NULL : fn(screenBounds, listOption, windowID, imageOption);
}

@interface NoShotSelectionView : NSView
@property NSPoint start;
@property NSPoint current;
@property BOOL dragging;
@end

@implementation NoShotSelectionView
- (BOOL)isFlipped { return NO; }
- (void)drawRect:(NSRect)dirtyRect {
	[[NSColor colorWithCalibratedWhite:0 alpha:0.22] setFill];
	NSRectFill(self.bounds);
	if (self.dragging) {
		NSRect rect = NSMakeRect(
			MIN(self.start.x, self.current.x),
			MIN(self.start.y, self.current.y),
			fabs(self.current.x - self.start.x),
			fabs(self.current.y - self.start.y)
		);
		[[NSColor colorWithCalibratedRed:0 green:0.34 blue:0.70 alpha:0.28] setFill];
		NSRectFillUsingOperation(rect, NSCompositingOperationSourceOver);
		[[NSColor whiteColor] setStroke];
		NSFrameRectWithWidth(rect, 2);
	}
}
@end

static int noshot_write_cgimage(CGImageRef image, const char *path) {
	if (image == NULL) {
		return 1;
	}
	@autoreleasepool {
		NSString *filePath = [NSString stringWithUTF8String:path];
		NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithCGImage:image];
		NSData *data = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
		if (data == nil) {
			return 2;
		}
		return [data writeToFile:filePath atomically:YES] ? 0 : 3;
	}
}

static CGRect noshot_cocoa_rect_to_cg(NSRect rect) {
	CGDirectDisplayID display = CGMainDisplayID();
	size_t height = CGDisplayPixelsHigh(display);
	return CGRectMake(rect.origin.x, height - rect.origin.y - rect.size.height, rect.size.width, rect.size.height);
}

static NSRect noshot_select_rect(void) {
	__block NSRect selected = NSZeroRect;

	void (^work)(void) = ^{
		@autoreleasepool {
			[NSApp activateIgnoringOtherApps:YES];
			NSScreen *screen = [NSScreen mainScreen];
			NSRect frame = [screen frame];
			NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
			                                               styleMask:NSWindowStyleMaskBorderless
			                                                 backing:NSBackingStoreBuffered
			                                                   defer:NO];
			[window setLevel:NSScreenSaverWindowLevel];
			[window setOpaque:NO];
			[window setBackgroundColor:[NSColor clearColor]];
			[window setIgnoresMouseEvents:NO];
			[window setReleasedWhenClosed:NO];
			[window makeKeyAndOrderFront:nil];

			NoShotSelectionView *view = [[NoShotSelectionView alloc] initWithFrame:NSMakeRect(0, 0, frame.size.width, frame.size.height)];
			[window setContentView:view];

			BOOL done = NO;
			while (!done) {
				NSEvent *event = [NSApp nextEventMatchingMask:(NSEventMaskLeftMouseDown | NSEventMaskLeftMouseDragged | NSEventMaskLeftMouseUp | NSEventMaskKeyDown)
				                                    untilDate:[NSDate distantFuture]
				                                       inMode:NSEventTrackingRunLoopMode
				                                      dequeue:YES];
				if (event.type == NSEventTypeKeyDown && event.keyCode == 53) {
					selected = NSZeroRect;
					done = YES;
				} else if (event.type == NSEventTypeLeftMouseDown) {
					view.start = [event locationInWindow];
					view.current = view.start;
					view.dragging = YES;
					[view setNeedsDisplay:YES];
				} else if (event.type == NSEventTypeLeftMouseDragged) {
					view.current = [event locationInWindow];
					[view setNeedsDisplay:YES];
				} else if (event.type == NSEventTypeLeftMouseUp) {
					view.current = [event locationInWindow];
					NSRect local = NSMakeRect(
						MIN(view.start.x, view.current.x),
						MIN(view.start.y, view.current.y),
						fabs(view.current.x - view.start.x),
						fabs(view.current.y - view.start.y)
					);
					if (local.size.width >= 2 && local.size.height >= 2) {
						selected = NSMakeRect(frame.origin.x + local.origin.x, frame.origin.y + local.origin.y, local.size.width, local.size.height);
					}
					done = YES;
				}
			}
			[window orderOut:nil];
		}
	};

	if ([NSThread isMainThread]) {
		work();
	} else {
		dispatch_sync(dispatch_get_main_queue(), work);
	}

	return selected;
}

static NSPoint noshot_select_point(void) {
	__block NSPoint point = NSMakePoint(NAN, NAN);

	void (^work)(void) = ^{
		@autoreleasepool {
			[NSApp activateIgnoringOtherApps:YES];
			NSScreen *screen = [NSScreen mainScreen];
			NSRect frame = [screen frame];
			NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
			                                               styleMask:NSWindowStyleMaskBorderless
			                                                 backing:NSBackingStoreBuffered
			                                                   defer:NO];
			[window setLevel:NSScreenSaverWindowLevel];
			[window setOpaque:NO];
			[window setBackgroundColor:[NSColor colorWithCalibratedWhite:0 alpha:0.12]];
			[window setIgnoresMouseEvents:NO];
			[window setReleasedWhenClosed:NO];
			[window makeKeyAndOrderFront:nil];

			BOOL done = NO;
			while (!done) {
				NSEvent *event = [NSApp nextEventMatchingMask:(NSEventMaskLeftMouseUp | NSEventMaskKeyDown)
				                                    untilDate:[NSDate distantFuture]
				                                       inMode:NSEventTrackingRunLoopMode
				                                      dequeue:YES];
				if (event.type == NSEventTypeKeyDown && event.keyCode == 53) {
					done = YES;
				} else if (event.type == NSEventTypeLeftMouseUp) {
					NSPoint local = [event locationInWindow];
					point = NSMakePoint(frame.origin.x + local.x, frame.origin.y + local.y);
					done = YES;
				}
			}
			[window orderOut:nil];
		}
	};

	if ([NSThread isMainThread]) {
		work();
	} else {
		dispatch_sync(dispatch_get_main_queue(), work);
	}

	return point;
}

static int noshot_capture_fullscreen(const char *path) {
	CGImageRef image = noshot_display_create_image(CGMainDisplayID());
	int status = noshot_write_cgimage(image, path);
	if (image != NULL) {
		CGImageRelease(image);
	}
	return status;
}

static int noshot_capture_region(const char *path) {
	NSRect cocoaRect = noshot_select_rect();
	if (NSIsEmptyRect(cocoaRect)) {
		return 10;
	}
	CGRect cgRect = noshot_cocoa_rect_to_cg(cocoaRect);
	CGImageRef image = noshot_display_create_image_for_rect(CGMainDisplayID(), cgRect);
	int status = noshot_write_cgimage(image, path);
	if (image != NULL) {
		CGImageRelease(image);
	}
	return status;
}

static int noshot_capture_window(const char *path) {
	NSPoint point = noshot_select_point();
	if (isnan(point.x) || isnan(point.y)) {
		return 10;
	}

	CGDirectDisplayID display = CGMainDisplayID();
	CGPoint cgPoint = CGPointMake(point.x, CGDisplayPixelsHigh(display) - point.y);
	CFArrayRef windows = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
	if (windows == NULL) {
		return 11;
	}

	CGWindowID selectedWindow = kCGNullWindowID;
	CFIndex count = CFArrayGetCount(windows);
	for (CFIndex i = 0; i < count; i++) {
		NSDictionary *info = (__bridge NSDictionary *)CFArrayGetValueAtIndex(windows, i);
		NSNumber *layer = info[(id)kCGWindowLayer];
		NSNumber *windowID = info[(id)kCGWindowNumber];
		NSDictionary *bounds = info[(id)kCGWindowBounds];
		if (layer == nil || windowID == nil || bounds == nil || [layer intValue] != 0) {
			continue;
		}
		CGRect rect;
		if (!CGRectMakeWithDictionaryRepresentation((__bridge CFDictionaryRef)bounds, &rect)) {
			continue;
		}
		if (CGRectContainsPoint(rect, cgPoint)) {
			selectedWindow = (CGWindowID)[windowID unsignedIntValue];
			break;
		}
	}
	CFRelease(windows);

	if (selectedWindow == kCGNullWindowID) {
		return 12;
	}

	CGImageRef image = noshot_window_list_create_image(CGRectNull, kCGWindowListOptionIncludingWindow, selectedWindow, kCGWindowImageBoundsIgnoreFraming);
	int status = noshot_write_cgimage(image, path);
	if (image != NULL) {
		CGImageRelease(image);
	}
	return status;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func nativeCapture(path string, mode CaptureMode) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var status C.int
	switch mode {
	case CaptureRegion:
		status = C.noshot_capture_region(cPath)
	case CaptureWindow:
		status = C.noshot_capture_window(cPath)
	case CaptureFullscreen:
		status = C.noshot_capture_fullscreen(cPath)
	default:
		return fmt.Errorf("unknown capture mode %d", mode)
	}
	if status == 10 {
		return fmt.Errorf("capture cancelled")
	}
	if status != 0 {
		return fmt.Errorf("native capture failed with status %d", int(status))
	}
	return nil
}
