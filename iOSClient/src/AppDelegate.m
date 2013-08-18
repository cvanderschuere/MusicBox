//
//  AppDelegate.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "AppDelegate.h"

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
}
- (void) onClose:(int)code reason:(NSString *)reason{
    NSLog(@"Websocket closed with reason: %@",reason);
    UIAlertView *errorView = [[UIAlertView alloc] initWithTitle:@"Websocket Error" message:reason delegate:nil cancelButtonTitle:@"Dismiss" otherButtonTitles: nil];
    [errorView show];
}
@end
