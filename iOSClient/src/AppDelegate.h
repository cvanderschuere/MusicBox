//
//  AppDelegate.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import <MDWamp/MDWamp.h>

//Global
const static NSString* baseURL = @"http://www.musicbox.com/";

@interface AppDelegate : UIResponder <UIApplicationDelegate,SPSessionDelegate,MDWampDelegate>

@property (strong, nonatomic) UIWindow *window;
@property (nonatomic, strong) NSOperationQueue *websocketRequestQueue;

@property (nonatomic, strong) MDWamp *ws;


@end
