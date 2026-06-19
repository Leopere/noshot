//go:build darwin && cgo

package app

/*
#cgo darwin CFLAGS: -x objective-c -fblocks
#cgo darwin LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#import <dispatch/dispatch.h>
#include <stdlib.h>

static int noshot_copy_image_to_clipboard(const char *path) {
	__block int status = 0;
	void (^work)(void) = ^{
	  @autoreleasepool {
		NSString *filePath = [NSString stringWithUTF8String:path];
		NSURL *fileURL = [NSURL fileURLWithPath:filePath];
		NSData *pngData = [NSData dataWithContentsOfFile:filePath];
		if (pngData == nil || [pngData length] == 0) {
			status = 1;
			return;
		}
		NSImage *image = [[NSImage alloc] initWithContentsOfFile:filePath];
		if (image == nil) {
			status = 2;
			return;
		}
		NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithData:pngData];
		NSData *tiffData = nil;
		if (rep != nil) {
			tiffData = [rep TIFFRepresentation];
		}

		NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
		NSInteger before = [pasteboard changeCount];
		[pasteboard clearContents];

		NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
		BOOL okPNG = [item setData:pngData forType:NSPasteboardTypePNG];
		BOOL okURL = [item setString:[fileURL absoluteString] forType:NSPasteboardTypeFileURL];
		BOOL okTIFF = YES;
		if (tiffData != nil) {
			okTIFF = [item setData:tiffData forType:NSPasteboardTypeTIFF];
		}
		BOOL okWrite = [pasteboard writeObjects:@[item]];
		NSInteger after = [pasteboard changeCount];
		if (!okPNG || !okURL || !okTIFF || !okWrite) {
			status = 3;
			return;
		}
		if (after == before) {
			status = 4;
			return;
		}
		NSLog(@"NoShot clipboard updated changeCount=%ld types=%@", (long)after, [pasteboard types]);
		status = 0;
	  }
	};
	if ([NSThread isMainThread]) {
		work();
	} else {
		dispatch_sync(dispatch_get_main_queue(), work);
	}
	return status;
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
