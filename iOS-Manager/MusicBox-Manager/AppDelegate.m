//
//  AppDelegate.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/14/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "AppDelegate.h"
#define DEBUG_MESSAGES YES

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
    
    // if you want debug log set this to YES, default is NO
    [MDWamp setDebug:DEBUG_MESSAGES];
    
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:[NSURL URLWithString:@"ws://ClientBalencer-394863257.us-west-2.elb.amazonaws.com:8080"]];
    
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
}

#pragma mark MDWamp Delegate
- (void) onOpen{
    [self.websocketRequestQueue setSuspended:NO];
    NSLog(@"Websocket is open");
}
- (void) onClose:(int)code reason:(NSString *)reason{
    NSLog(@"Websocket closed with reason: %@",reason);
    
    if (DEBUG_MESSAGES) {
        UIAlertView *errorView = [[UIAlertView alloc] initWithTitle:@"Websocket Error" message:reason delegate:nil cancelButtonTitle:@"Dismiss" otherButtonTitles: nil];
        [errorView show];
    }
}

- (void)authenticateWebsocketWithUsername:(NSString*)username Password:(NSString*)password Callback:(void(^)(User* user, NSError* error))callback{
    //Lookup sessionid for username/password
    [self.ws call:[NSString stringWithFormat:@"%@user/startSession",baseURL] success:^(NSString *callURI, id result) {
        NSDictionary* resultDict = (NSDictionary*) result;
        NSString *sessionID = resultDict[@"sessionID"];
        
        //Authenticate websocket
        [self.ws authWithKey:username Secret:sessionID Extra:nil Success:^(NSString *answer) {
            NSLog(@"Authenticated");
            User* newUser = [[User alloc] init];
            newUser.username = username;
            newUser.sessionID = sessionID;
            callback(newUser,nil);
        } Error:^(NSString *procCall, NSString *errorURI, NSString *errorDetails) {
            NSLog(@"Auth Fail:%@ %@",procCall,errorDetails);
            callback(nil,[NSError errorWithDomain:errorURI code:0 userInfo:nil]);
        }];
        
    } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
        callback(nil,[NSError errorWithDomain:@"Registration Error" code:500 userInfo:nil]);
    } args:@{
        @"username": username,
        @"password": password
     }, nil];
    
}




@end
