//
//  Device.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "Device.h"

@implementation Device

+ (instancetype) deviceWithDict:(NSDictionary*)dict{
    Device *newDevice = [[Device alloc] init];
    newDevice.tracks = [NSArray array];
    
    //Load from dict
    newDevice.deviceName = dict[@"DeviceName"];
    newDevice.identifier = dict[@"ID"];
    newDevice.theme = dict[@"Theme"];
    newDevice.user = dict[@"User"];
    
    NSArray *coords = dict[@"Location"];
    newDevice.location = [[CLLocation alloc] initWithLatitude:[coords[0] doubleValue] longitude:[coords[1] doubleValue]];
    
    return newDevice;
}


@end
