#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>

extern void goHandleHotkey(int id);
extern void goHandleMenu(int action);

static NSStatusItem *statusItem;
static NSObject *menuTarget;
static EventHotKeyRef hotKeyRefs[5];

@interface NoShotMenuTarget : NSObject
- (void)captureRegion:(id)sender;
- (void)captureWindow:(id)sender;
- (void)captureFullscreen:(id)sender;
- (void)codexExplain:(id)sender;
- (void)codexCustom:(id)sender;
- (void)openScreenshots:(id)sender;
- (void)editConfig:(id)sender;
- (void)captureSelfTest:(id)sender;
- (void)codexSelfTest:(id)sender;
- (void)quit:(id)sender;
@end

@implementation NoShotMenuTarget
- (void)captureRegion:(id)sender { goHandleHotkey(1); }
- (void)captureWindow:(id)sender { goHandleHotkey(2); }
- (void)captureFullscreen:(id)sender { goHandleHotkey(3); }
- (void)codexExplain:(id)sender { goHandleHotkey(4); }
- (void)codexCustom:(id)sender { goHandleHotkey(5); }
- (void)openScreenshots:(id)sender { goHandleMenu(1); }
- (void)editConfig:(id)sender { goHandleMenu(2); }
- (void)captureSelfTest:(id)sender { goHandleMenu(3); }
- (void)codexSelfTest:(id)sender { goHandleMenu(4); }
- (void)quit:(id)sender { [NSApp terminate:nil]; }
@end

static OSStatus hotKeyHandler(EventHandlerCallRef nextHandler, EventRef event, void *userData) {
	EventHotKeyID hotKeyID;
	GetEventParameter(
		event,
		kEventParamDirectObject,
		typeEventHotKeyID,
		NULL,
		sizeof(hotKeyID),
		NULL,
		&hotKeyID
	);
	goHandleHotkey((int)hotKeyID.id);
	return noErr;
}

static void registerHotKey(UInt32 keyCode, UInt32 identifier) {
	EventHotKeyID hotKeyID;
	hotKeyID.signature = 'NSHT';
	hotKeyID.id = identifier;
	RegisterEventHotKey(
		keyCode,
		cmdKey | shiftKey,
		hotKeyID,
		GetApplicationEventTarget(),
		0,
		&hotKeyRefs[identifier - 1]
	);
}

static NSMenuItem *menuItem(NSString *title, SEL action, NSString *keyEquivalent) {
	NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:action keyEquivalent:keyEquivalent];
	[item setTarget:menuTarget];
	return item;
}

void noshot_run(void) {
	@autoreleasepool {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

		EventTypeSpec eventType;
		eventType.eventClass = kEventClassKeyboard;
		eventType.eventKind = kEventHotKeyPressed;
		InstallApplicationEventHandler(&hotKeyHandler, 1, &eventType, NULL, NULL);

		menuTarget = [[NoShotMenuTarget alloc] init];
		statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
		NSImage *statusImage = [NSImage imageNamed:@"noshot-status"];
		if (statusImage != nil) {
			[statusImage setTemplate:YES];
			[statusImage setSize:NSMakeSize(18, 18)];
			[[statusItem button] setImage:statusImage];
			[[statusItem button] setToolTip:@"NoShot"];
		} else {
			[[statusItem button] setTitle:@"NoShot"];
		}

		NSMenu *menu = [[NSMenu alloc] initWithTitle:@"NoShot"];
		[menu addItem:menuItem(@"Capture Region", @selector(captureRegion:), @"1")];
		[menu addItem:menuItem(@"Capture Window", @selector(captureWindow:), @"2")];
		[menu addItem:menuItem(@"Capture Fullscreen", @selector(captureFullscreen:), @"3")];
		[menu addItem:[NSMenuItem separatorItem]];
		[menu addItem:menuItem(@"Explain Screenshot with Codex", @selector(codexExplain:), @"4")];
		[menu addItem:menuItem(@"Ask Codex with Custom Prompt", @selector(codexCustom:), @"5")];
		[menu addItem:[NSMenuItem separatorItem]];
		[menu addItem:menuItem(@"Open Screenshots Folder", @selector(openScreenshots:), @"")];
		[menu addItem:menuItem(@"Edit Config", @selector(editConfig:), @"")];
		[menu addItem:[NSMenuItem separatorItem]];
		[menu addItem:menuItem(@"Run Capture Self-Test", @selector(captureSelfTest:), @"")];
		[menu addItem:menuItem(@"Run Codex Self-Test", @selector(codexSelfTest:), @"")];
		[menu addItem:[NSMenuItem separatorItem]];
		[menu addItem:menuItem(@"Quit NoShot", @selector(quit:), @"q")];
		[statusItem setMenu:menu];

		registerHotKey(kVK_ANSI_1, 1);
		registerHotKey(kVK_ANSI_2, 2);
		registerHotKey(kVK_ANSI_3, 3);
		registerHotKey(kVK_ANSI_4, 4);
		registerHotKey(kVK_ANSI_5, 5);

		[NSApp run];
	}
}
