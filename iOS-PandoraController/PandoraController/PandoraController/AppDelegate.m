//
//  AppDelegate.m
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import "AppDelegate.h"

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions
{
    // Override point for customization after application launch.
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
-(void) connectToWebSocket{
    if (self.ws) {
        [self.ws disconnect];
        self.ws = nil;
        [self performSelector:@selector(connectToWebSocket) withObject:nil afterDelay:1];
        return;
    }
    
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:[NSURL URLWithString:@"ws://ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"]];
    
    //Set delegate to receive open & close info
    self.ws = [[MBWebsocket alloc] initWithURLRequest:request delegate:self];
    self.ws.baseURL = @"http://www.musicbox.com/";
    
    // set if MDWAMP should automatically try to reconnect after a network fail default YES
    [self.ws setShouldAutoreconnect:YES];
    
    // set number of times it tries to autoreconnect after a fail
    [self.ws setAutoreconnectMaxRetries:20];
    
    // set seconds between each reconnection try
    [self.ws setAutoreconnectDelay:1];
    
    //Actually connect
    [self.ws connect];
}

#pragma mark - MDWamp Delegate
- (void) onOpen{
    [self.ws.requestQueue setSuspended:NO];
    
    self.pingTimer = [NSTimer timerWithTimeInterval:30 target:self selector:@selector(pingWebsocket) userInfo:nil repeats:YES];
    [[NSRunLoop currentRunLoop] addTimer:self.pingTimer forMode:NSDefaultRunLoopMode];
}

- (void) onClose:(NSInteger)code reason:(NSString *)reason{
    NSLog(@"Websocket closed with reason: %@",reason);
    
    if (false) {
        UIAlertView *errorView = [[UIAlertView alloc] initWithTitle:@"Websocket Error" message:reason delegate:nil cancelButtonTitle:@"Dismiss" otherButtonTitles: nil];
        [errorView show];
    }
}

- (void) pingWebsocket{
    if (self.ws.isConnected) {
        //Not an actual method but will keep the websocket open
        [self.ws publish:@"ping" toTopic:[self.ws.baseURL stringByAppendingString:@"ping"]];
    }
}


@end
