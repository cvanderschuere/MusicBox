//
//  AppDelegate.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import <MDWamp/MDWamp.h>
#import "TestFlight.h"

//Global (used in other parts of the app)
const static NSString* baseURL = @"http://www.musicbox.com/";

@interface AppDelegate : UIResponder <UIApplicationDelegate,MDWampDelegate>

@property (strong, nonatomic) UIWindow *window;
@property (nonatomic, strong) NSOperationQueue *websocketRequestQueue;

@property (nonatomic, strong) MDWamp *ws;


@end
