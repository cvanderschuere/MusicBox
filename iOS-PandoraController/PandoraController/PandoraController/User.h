//
//  User.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "Device.h"

@interface User : NSObject


@property (nonatomic,strong) NSString* username;
@property (nonatomic,strong) NSString* sessionID;

@property (nonatomic,strong) NSArray* stations;
@property (nonatomic,strong) NSArray* devices;

@property (nonatomic,strong) Device *selectedDevice;


//Permissions
@property (nonatomic,strong) NSArray* rpcPerms;
@property (nonatomic,strong) NSDictionary* pubSubPerms;

@end
