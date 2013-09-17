//
//  AppDelegate.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/14/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "User.h"

const static NSString* baseURL = @"http://www.musicbox.com/";

@interface AppDelegate : UIResponder <UIApplicationDelegate,MDWampDelegate>

@property (strong, nonatomic) UIWindow *window;

@property (nonatomic, strong) NSOperationQueue *websocketRequestQueue;
@property (nonatomic, strong) MDWamp *ws;



#pragma mark - Websocket methods
- (void)authenticateWebsocketWithUsername:(NSString*)username Password:(NSString*)password Callback:(void(^)(User* newUser,NSError* error))callback;

@end
