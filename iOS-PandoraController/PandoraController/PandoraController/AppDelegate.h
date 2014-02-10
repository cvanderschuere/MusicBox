//
//  AppDelegate.h
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>

@interface AppDelegate : UIResponder <UIApplicationDelegate, MDWampClientDelegate>

@property (strong, nonatomic) UIWindow *window;

@property (nonatomic, strong) MBWebsocket *ws;
@property (nonatomic, strong) NSTimer* pingTimer;

-(void) connectToWebSocket;

@end
