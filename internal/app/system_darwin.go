//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c -fblocks
#cgo darwin LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#import <dispatch/dispatch.h>
#include <stdlib.h>

static NSString *noshot_string(const char *value) {
	if (value == NULL) {
		return @"";
	}
	return [NSString stringWithUTF8String:value];
}

static void noshot_notify(const char *title, const char *message) {
	@autoreleasepool {
		NSLog(@"%@: %@", noshot_string(title), noshot_string(message));
	}
}

static char *noshot_ask_custom_prompt(void) {
	__block char *result = NULL;

	void (^work)(void) = ^{
		@autoreleasepool {
			[NSApp activateIgnoringOtherApps:YES];

			NSAlert *alert = [[NSAlert alloc] init];
			alert.messageText = @"Ask Codex";
			alert.informativeText = @"Ask Codex about this screenshot:";
			[alert addButtonWithTitle:@"Run"];
			[alert addButtonWithTitle:@"Cancel"];

			NSTextField *input = [[NSTextField alloc] initWithFrame:NSMakeRect(0, 0, 420, 24)];
			[input setStringValue:@""];
			alert.accessoryView = input;

			NSWindow *window = [alert window];
			[window setInitialFirstResponder:input];
			NSInteger response = [alert runModal];
			if (response == NSAlertFirstButtonReturn) {
				const char *utf8 = [[input stringValue] UTF8String];
				if (utf8 != NULL) {
					result = strdup(utf8);
				}
			}
		}
	};

	if ([NSThread isMainThread]) {
		work();
	} else {
		dispatch_sync(dispatch_get_main_queue(), work);
	}

	return result;
}

static int noshot_copy_text_to_clipboard(const char *text) {
	@autoreleasepool {
		NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
		[pasteboard clearContents];
		BOOL ok = [pasteboard setString:noshot_string(text) forType:NSPasteboardTypeString];
		return ok ? 0 : 1;
	}
}

static int noshot_open_path(const char *path) {
	@autoreleasepool {
		NSURL *url = [NSURL fileURLWithPath:noshot_string(path)];
		BOOL ok = [[NSWorkspace sharedWorkspace] openURL:url];
		return ok ? 0 : 1;
	}
}

static int noshot_edit_path(const char *path) {
	return noshot_open_path(path);
}
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"unsafe"
)

func Notify(title, message string) {
	cTitle := C.CString(title)
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMessage))
	C.noshot_notify(cTitle, cMessage)
}

func AskCustomPrompt(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	result := C.noshot_ask_custom_prompt()
	if result == nil {
		return "", fmt.Errorf("custom prompt cancelled")
	}
	defer C.free(unsafe.Pointer(result))

	return strings.TrimSpace(C.GoString(result)), nil
}

func OpenPath(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if status := C.noshot_open_path(cPath); status != 0 {
		return fmt.Errorf("could not open %s", path)
	}
	return nil
}

func EditPath(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if status := C.noshot_edit_path(cPath); status != 0 {
		return fmt.Errorf("could not edit %s", path)
	}
	return nil
}

func CopyTextToClipboard(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	if status := C.noshot_copy_text_to_clipboard(cText); status != 0 {
		return fmt.Errorf("pasteboard rejected text with status %d", int(status))
	}
	return nil
}
