//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c -fblocks
#cgo darwin LDFLAGS: -framework AppKit -framework CoreGraphics -framework ScreenCaptureKit
#import <AppKit/AppKit.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>
#include <string.h>
#include <stdlib.h>

static char noshot_last_error[1024] = "";

static void noshot_clear_error(void) {
	noshot_last_error[0] = '\0';
}

static void noshot_set_error(NSString *message) {
	if (message == nil) {
		noshot_last_error[0] = '\0';
		return;
	}
	snprintf(noshot_last_error, sizeof(noshot_last_error), "%s", [message UTF8String]);
}

static char *noshot_last_error_copy(void) {
	return strdup(noshot_last_error);
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
		noshot_set_error(@"ScreenCaptureKit returned an empty image");
		return 1;
	}
	@autoreleasepool {
		NSString *filePath = [NSString stringWithUTF8String:path];
		NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithCGImage:image];
		NSData *data = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
		if (data == nil) {
			noshot_set_error(@"Could not encode captured image as PNG");
			return 2;
		}
		if (![data writeToFile:filePath atomically:YES]) {
			noshot_set_error([NSString stringWithFormat:@"Could not write PNG to %@", filePath]);
			return 3;
		}
		return 0;
	}
}

static NSScreen *noshot_screen_for_display_id(CGDirectDisplayID displayID) {
	for (NSScreen *screen in [NSScreen screens]) {
		NSNumber *screenNumber = screen.deviceDescription[@"NSScreenNumber"];
		if (screenNumber != nil && [screenNumber unsignedIntValue] == displayID) {
			return screen;
		}
	}
	return [NSScreen mainScreen];
}

static SCDisplay *noshot_display_for_cocoa_rect(SCShareableContent *content, NSRect rect) {
	NSPoint center = NSMakePoint(NSMidX(rect), NSMidY(rect));
	for (SCDisplay *display in content.displays) {
		NSScreen *screen = noshot_screen_for_display_id(display.displayID);
		if (screen != nil && NSPointInRect(center, screen.frame)) {
			return display;
		}
	}
	return [content.displays firstObject];
}

static SCDisplay *noshot_display_for_cocoa_point(SCShareableContent *content, NSPoint point) {
	for (SCDisplay *display in content.displays) {
		NSScreen *screen = noshot_screen_for_display_id(display.displayID);
		if (screen != nil && NSPointInRect(point, screen.frame)) {
			return display;
		}
	}
	return [content.displays firstObject];
}

static CGRect noshot_source_rect_for_display(SCDisplay *display, NSRect cocoaRect) {
	NSScreen *screen = noshot_screen_for_display_id(display.displayID);
	NSRect screenFrame = screen.frame;
	CGFloat x = cocoaRect.origin.x - screenFrame.origin.x;
	CGFloat yFromBottom = cocoaRect.origin.y - screenFrame.origin.y;
	CGFloat y = screenFrame.size.height - yFromBottom - cocoaRect.size.height;
	return CGRectMake(MAX(0, x), MAX(0, y), cocoaRect.size.width, cocoaRect.size.height);
}

static CGPoint noshot_sck_point_for_display(SCDisplay *display, NSPoint cocoaPoint) {
	NSScreen *screen = noshot_screen_for_display_id(display.displayID);
	NSRect screenFrame = screen.frame;
	CGFloat x = cocoaPoint.x - screenFrame.origin.x + display.frame.origin.x;
	CGFloat yFromBottom = cocoaPoint.y - screenFrame.origin.y;
	CGFloat y = screenFrame.size.height - yFromBottom + display.frame.origin.y;
	return CGPointMake(x, y);
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
			noshot_set_error(error.localizedDescription ?: @"ScreenCaptureKit could not enumerate shareable content");
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
	config.capturesAudio = NO;
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
			noshot_set_error(error.localizedDescription ?: @"ScreenCaptureKit screenshot capture failed");
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
	noshot_clear_error();
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
	noshot_clear_error();
	NSRect selectedRect = noshot_select_rect();
	if (NSIsEmptyRect(selectedRect)) {
		noshot_set_error(@"Capture cancelled");
		return 10;
	}

	int status = 0;
	SCShareableContent *content = noshot_shareable_content(&status);
	if (status != 0) {
		return status;
	}
	SCDisplay *display = noshot_display_for_cocoa_rect(content, selectedRect);
	if (display == nil) {
		noshot_set_error(@"No display matched the selected region");
		return 22;
	}

	SCContentFilter *filter = [[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];
	CGRect sourceRect = noshot_source_rect_for_display(display, selectedRect);
	return noshot_capture_filter(filter, noshot_display_config(display, sourceRect), path);
}

static int noshot_capture_window(const char *path) {
	noshot_clear_error();
	NSPoint point = noshot_select_point();
	if (isnan(point.x) || isnan(point.y)) {
		noshot_set_error(@"Capture cancelled");
		return 10;
	}

	int status = 0;
	SCShareableContent *content = noshot_shareable_content(&status);
	if (status != 0) {
		return status;
	}

	SCDisplay *clickedDisplay = noshot_display_for_cocoa_point(content, point);
	if (clickedDisplay == nil) {
		noshot_set_error(@"No display matched the clicked point");
		return 22;
	}
	CGPoint sckPoint = noshot_sck_point_for_display(clickedDisplay, point);
	SCWindow *selectedWindow = nil;
	for (SCWindow *window in content.windows) {
		if (!window.isOnScreen || window.windowLayer != 0) {
			continue;
		}
		if (CGRectContainsPoint(window.frame, sckPoint)) {
			selectedWindow = window;
			break;
		}
	}
	if (selectedWindow == nil) {
		noshot_set_error(@"No window found under the clicked point");
		return 12;
	}

	SCContentFilter *filter = [[SCContentFilter alloc] initWithDesktopIndependentWindow:selectedWindow];
	SCShareableContentInfo *info = [SCShareableContent infoForFilter:filter];
	CGRect contentRect = info.contentRect;
	CGFloat scale = MAX((CGFloat)info.pointPixelScale, 1.0);
	SCStreamConfiguration *config = [[SCStreamConfiguration alloc] init];
	config.width = (size_t)MAX(1.0, contentRect.size.width * scale);
	config.height = (size_t)MAX(1.0, contentRect.size.height * scale);
	config.pixelFormat = 'BGRA';
	config.showsCursor = NO;
	config.capturesAudio = NO;
	config.queueDepth = 1;
	config.ignoreShadowsSingleWindow = YES;
	return noshot_capture_filter(filter, config, path);
}
*/
import "C"

import (
	"fmt"
	"strings"
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
		detail := lastNativeCaptureError()
		if detail != "" {
			Logf("native capture failed status=%d detail=%s", int(status), detail)
			return fmt.Errorf("%s", detail)
		}
		Logf("native capture failed status=%d", int(status))
		return fmt.Errorf("native capture failed with status %d", int(status))
	}
	return nil
}

func lastNativeCaptureError() string {
	cMessage := C.noshot_last_error_copy()
	if cMessage == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cMessage))
	return strings.TrimSpace(C.GoString(cMessage))
}
