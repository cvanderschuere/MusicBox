//
//  AppDelegate.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "AppDelegate.h"
#import "appkey.h"

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions
{
    // Override point for customization after application launch.
    
    //Connect to websocket
    // if you want debug log set this to YES, default is NO
    [MDWamp setDebug:YES];
    
    self.ws = [[MDWamp alloc] initWithUrl:@"ws://ec2-54-218-97-11.us-west-2.compute.amazonaws.com:8080" delegate:self];
    
    // set if MDWAMP should automatically try to reconnect after a network fail default YES
    [self.ws setShouldAutoreconnect:YES];
    
    // set number of times it tries to autoreconnect after a fail
    [self.ws setAutoreconnectMaxRetries:2];
    
    // set seconds between each reconnection try
    [self.ws setAutoreconnectDelay:5];
    
    //Create Request Queue
    self.requestQueue = [[NSOperationQueue alloc] init];
    [self.requestQueue setSuspended:YES];
    
    //Actually connect
    [self.ws connect];
    
    NSString *userAgent = [[[NSBundle mainBundle] infoDictionary] valueForKey:(__bridge NSString *)kCFBundleIdentifierKey];
	NSData *appKey = [NSData dataWithBytes:&g_appkey length:g_appkey_size];
    
	NSError *error = nil;
	[SPSession initializeSharedSessionWithApplicationKey:appKey
											   userAgent:userAgent
										   loadingPolicy:SPAsyncLoadingManual
												   error:&error];
	if (error != nil) {
		NSLog(@"CocoaLibSpotify init failed: %@", error);
		abort();
	}
	[[SPSession sharedSession] setDelegate:self];
    
    //Login to spotify
    NSUserDefaults *defaults = [NSUserDefaults standardUserDefaults];
    NSString *username = [defaults valueForKey:@"userName"];
    
    if (username)
      [[SPSession sharedSession] attemptLoginWithUserName:username existingCredential:[defaults valueForKey:@"credential"]];
    else{
        [self performSelector:@selector(showLogin) withObject:NULL afterDelay:0.0]; //Show after this method finishes
    }
    
        
    return YES;
}
							
- (void)applicationWillResignActive:(UIApplication *)application
{
    // Sent when the application is about to move from active to inactive state. This can occur for certain types of temporary interruptions (such as an incoming phone call or SMS message) or when the user quits the application and it begins the transition to the background state.
    // Use this method to pause ongoing tasks, disable timers, and throttle down OpenGL ES frame rates. Games should use this method to pause the game.
}

- (void)applicationDidEnterBackground:(UIApplication *)application
{
    // Use this method to release shared resources, save user data, invalidate timers, and store enough application state information to restore your application to its current state in case it is terminated later. 
    // If your application supports background execution, this method is called instead of applicationWillTerminate: when the user quits.
}

- (void)applicationWillEnterForeground:(UIApplication *)application
{
    // Called as part of the transition from the background to the inactive state; here you can undo many of the changes made on entering the background.
}

- (void)applicationDidBecomeActive:(UIApplication *)application
{
    // Restart any tasks that were paused (or not yet started) while the application was inactive. If the application was previously in the background, optionally refresh the user interface.
}

- (void)applicationWillTerminate:(UIApplication *)application
{
    // Called when the application is about to terminate. Save data if appropriate. See also applicationDidEnterBackground:.
}

#pragma mark SPSession Delegate
-(UIViewController *)viewControllerToPresentLoginViewForSession:(SPSession *)aSession {
    //All methods needing this viewcontroller use this method so you only need to change this
    return self.window.rootViewController;
}
-(void)session:(SPSession *)aSession didGenerateLoginCredentials:(NSString *)credential forUserName:(NSString *)userName {
	
	NSUserDefaults *defaults = [NSUserDefaults standardUserDefaults];    
	[defaults setValue:credential forKey:@"credential"];
	[defaults setValue:userName forKey:@"userName"];
}
- (void)sessionDidLoginSuccessfully:(SPSession *)aSession{
}
- (void)session:(SPSession *)aSession didFailToLoginWithError:(NSError *)error{
    [self showLogin];
}
-(void) session:(SPSession *)aSession didEncounterNetworkError:(NSError *)error{
    NSLog(@"Error: %@",error);
}
-(void) session:(SPSession *)aSession didEncounterScrobblingError:(NSError *)error{
    NSLog(@"Error: %@",error);
}
-(void)sessionDidLogOut:(SPSession *)aSession {
    [self showLogin];
}
-(void) showLogin{
    if ([self viewControllerToPresentLoginViewForSession:[SPSession sharedSession]].presentedViewController != nil) return;
    SPLoginViewController *controller = [SPLoginViewController loginControllerForSession:[SPSession sharedSession]];
	controller.allowsCancel = NO;
	[[self viewControllerToPresentLoginViewForSession:[SPSession sharedSession]] presentViewController:controller animated:YES completion:NULL];
}

-(void)session:(SPSession *)aSession didLogMessage:(NSString *)aMessage; {}
-(void)sessionDidChangeMetadata:(SPSession *)aSession; {}

-(void)session:(SPSession *)aSession recievedMessageForUser:(NSString *)aMessage; {
	return;
	UIAlertView *alert = [[UIAlertView alloc] initWithTitle:@"Message from Spotify"
													message:aMessage
												   delegate:nil
										  cancelButtonTitle:@"OK"
										  otherButtonTitles:nil];
	[alert show];
}

#pragma mark MDWamp Delegate
- (void) onOpen{
    NSLog(@"Websocket is open");
    [self.requestQueue setSuspended:NO];
}
- (void) onClose:(int)code reason:(NSString *)reason{
    NSLog(@"Websocket closed with reason: %@",reason);
}
@end
