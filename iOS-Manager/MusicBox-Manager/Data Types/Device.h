//
//  Device.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "Track.h"

@interface Device : NSObject
@property (nonatomic,strong) NSString* identifier;
@property (nonatomic,strong) NSString* user;
@property (nonatomic,strong) NSString* deviceName;

@property (nonatomic,strong) NSNumber *isPlaying; //yes (playing) / no (paused)

//Moment.us information
@property (nonatomic,strong) CLLocation* location;
@property (nonatomic,strong) NSString* theme;
@property (nonatomic,strong) NSDictionary* themeObj;

@property (nonatomic,strong) NSMutableArray* tracks; //array of Track*


+ (instancetype) deviceWithDict:(NSDictionary*)dict;

@end
