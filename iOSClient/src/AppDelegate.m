//
//  AppDelegate.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "AppDelegate.h"
#import "PlaylistViewController.h"

#define DEBUG_MESSAGES YES

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions
{
    // Override point for customization after application launch.
    
    //Setup Testflight
    [TestFlight takeOff:@"9bb727dd-5437-4e8b-b2eb-4a9508e74bb0"];
    
    //Connect to websocket
    // if you want debug log set this to YES, default is NO
    [MDWamp setDebug:DEBUG_MESSAGES];
    
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:[NSURL URLWithString:@"ws://ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"]];
    [request addValue:@"christopher.vanderschuere@gmail.com" forHTTPHeaderField:@"musicbox-username"];
    [request addValue:@"HryV3rtCBEBdvjW7fcTjKA" forHTTPHeaderField:@"musicbox-session-id"];
    
    self.ws = [[MDWamp alloc] initWithURLRequest:request delegate:self];
    
    // set if MDWAMP should automatically try to reconnect after a network fail default YES
    [self.ws setShouldAutoreconnect:YES];
    
    // set number of times it tries to autoreconnect after a fail
    [self.ws setAutoreconnectMaxRetries:10];
    
    // set seconds between each reconnection try
    [self.ws setAutoreconnectDelay:5];
    
    //Create Request Queue for websocket
    self.websocketRequestQueue = [[NSOperationQueue alloc] init];
    [self.websocketRequestQueue setSuspended:YES];
    
    //Actually connect
    [self.ws connect];
        
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


#pragma mark MDWamp Delegate
- (void) onOpen{
    NSLog(@"Websocket is open");
    [self.websocketRequestQueue setSuspended:NO];
    PlaylistViewController *playlistVC = (PlaylistViewController*) self.window.rootViewController;
    [playlistVC.refreshControl endRefreshing];

}
- (void) onClose:(int)code reason:(NSString *)reason{
    NSLog(@"Websocket closed with reason: %@",reason);
    
    if (DEBUG_MESSAGES) {
        UIAlertView *errorView = [[UIAlertView alloc] initWithTitle:@"Websocket Error" message:reason delegate:nil cancelButtonTitle:@"Dismiss" otherButtonTitles: nil];
        [errorView show];
    }
    
    PlaylistViewController *playlistVC = (PlaylistViewController*) self.window.rootViewController;
    [playlistVC.refreshControl endRefreshing];

}
@end
