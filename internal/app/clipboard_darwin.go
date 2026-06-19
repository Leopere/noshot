//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#include <stdlib.h>

static int noshot_copy_image_to_clipboard(const char *path) {
	@autoreleasepool {
		NSString *filePath = [NSString stringWithUTF8String:path];
		NSImage *image = [[NSImage alloc] initWithContentsOfFile:filePath];
		if (image == nil) {
			return 1;
		}
		NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
		[pasteboard clearContents];
		BOOL ok = [pasteboard writeObjects:@[image]];
		return ok ? 0 : 2;
	}
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func CopyImageToClipboard(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if status := C.noshot_copy_image_to_clipboard(cPath); status != 0 {
		return fmt.Errorf("pasteboard rejected image with status %d", int(status))
	}
	return nil
}
