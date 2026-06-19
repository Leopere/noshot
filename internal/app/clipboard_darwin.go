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
		NSURL *fileURL = [NSURL fileURLWithPath:filePath];
		NSData *pngData = [NSData dataWithContentsOfFile:filePath];
		if (pngData == nil || [pngData length] == 0) {
			return 1;
		}
		NSImage *image = [[NSImage alloc] initWithContentsOfFile:filePath];
		if (image == nil) {
			return 2;
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
			return 3;
		}
		if (after == before) {
			return 4;
		}
		return 0;
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
