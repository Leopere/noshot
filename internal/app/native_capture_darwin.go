//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c -fblocks
#cgo darwin LDFLAGS: -framework AppKit -framework CoreGraphics -framework ScreenCaptureKit
#import <AppKit/AppKit.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>
#include <stdlib.h>

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

static SCShareableContent *noshot_shareable_content(int *status) {
	__block SCShareableContent *result = nil;
	__block int localStatus = 0;
	dispatch_semaphore_t sema = dispatch_semaphore_create(0);
	[SCShareableContent getShareableContentExcludingDesktopWindows:NO onScreenWindowsOnly:YES completionHandler:^(SCShareableContent *content, NSError *error) {
		if (error != nil || content == nil) {
			localStatus = 20;
		} else {
			result = content;
		}
		dispatch_semaphore_signal(sema);
	}];
	dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
	if (status != NULL) {
		*status = localStatus;
	}
	return result;
}

static SCDisplay *noshot_main_display(SCShareableContent *content) {
	CGDirectDisplayID mainDisplay = CGMainDisplayID();
	for (SCDisplay *display in content.displays) {
		if (display.displayID == mainDisplay) {
			return display;
		}
	}
	return [content.displays firstObject];
}

static SCStreamConfiguration *noshot_display_config(SCDisplay *display, CGRect sourceRect) {
	SCStreamConfiguration *config = [[SCStreamConfiguration alloc] init];
	CGDirectDisplayID displayID = display.displayID;
	CGRect rect = CGRectIsEmpty(sourceRect) ? display.frame : sourceRect;
	CGFloat scaleX = (CGFloat)CGDisplayPixelsWide(displayID) / MAX((CGFloat)display.width, 1.0);
	CGFloat scaleY = (CGFloat)CGDisplayPixelsHigh(displayID) / MAX((CGFloat)display.height, 1.0);
	config.width = (size_t)MAX(1.0, rect.size.width * scaleX);
	config.height = (size_t)MAX(1.0, rect.size.height * scaleY);
	config.pixelFormat = 'BGRA';
	config.showsCursor = NO;
	config.queueDepth = 1;
	if (!CGRectIsEmpty(sourceRect)) {
		config.sourceRect = sourceRect;
	}
	return config;
}

static int noshot_capture_filter(SCContentFilter *filter, SCStreamConfiguration *config, const char *path) {
	__block int status = 0;
	dispatch_semaphore_t sema = dispatch_semaphore_create(0);
	[SCScreenshotManager captureImageWithFilter:filter configuration:config completionHandler:^(CGImageRef image, NSError *error) {
		if (error != nil) {
			status = 21;
		} else {
			status = noshot_write_cgimage(image, path);
		}
		dispatch_semaphore_signal(sema);
	}];
	dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
	return status;
}

static int noshot_capture_fullscreen(const char *path) {
	int status = 0;
	SCShareableContent *content = noshot_shareable_content(&status);
	if (status != 0) {
		return status;
	}
	SCDisplay *display = noshot_main_display(content);
	if (display == nil) {
		return 22;
	}
	SCContentFilter *filter = [[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];
	return noshot_capture_filter(filter, noshot_display_config(display, CGRectNull), path);
}

static int noshot_capture_region(const char *path) {
	NSRect selectedRect = noshot_select_rect();
	if (NSIsEmptyRect(selectedRect)) {
		return 10;
	}

	int status = 0;
	SCShareableContent *content = noshot_shareable_content(&status);
	if (status != 0) {
		return status;
	}
	SCDisplay *display = noshot_main_display(content);
	if (display == nil) {
		return 22;
	}

	SCContentFilter *filter = [[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];
	return noshot_capture_filter(filter, noshot_display_config(display, selectedRect), path);
}

static int noshot_capture_window(const char *path) {
	NSPoint point = noshot_select_point();
	if (isnan(point.x) || isnan(point.y)) {
		return 10;
	}

	int status = 0;
	SCShareableContent *content = noshot_shareable_content(&status);
	if (status != 0) {
		return status;
	}

	SCWindow *selectedWindow = nil;
	for (SCWindow *window in content.windows) {
		if (!window.isOnScreen || window.windowLayer != 0) {
			continue;
		}
		if (CGRectContainsPoint(window.frame, point)) {
			selectedWindow = window;
			break;
		}
	}
	if (selectedWindow == nil) {
		return 12;
	}

	SCContentFilter *filter = [[SCContentFilter alloc] initWithDesktopIndependentWindow:selectedWindow];
	SCStreamConfiguration *config = [[SCStreamConfiguration alloc] init];
	config.width = (size_t)MAX(1.0, selectedWindow.frame.size.width);
	config.height = (size_t)MAX(1.0, selectedWindow.frame.size.height);
	config.pixelFormat = 'BGRA';
	config.showsCursor = NO;
	config.queueDepth = 1;
	config.ignoreShadowsSingleWindow = YES;
	return noshot_capture_filter(filter, config, path);
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
