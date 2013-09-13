//
//  MusicBox.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 6/29/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface MusicBoxTrack : NSObject
@property (nonatomic,strong) NSString *service;
@property (nonatomic,strong) NSString *url;

@property (nonatomic,strong) NSString *trackName;
@property (nonatomic,strong) NSString *artistName;
@property (nonatomic,strong) NSString *albumName;

@property (nonatomic,strong) NSURL *artworkURL;

+(instancetype) trackWithService:(NSString*)serviceName Url:(NSString*)url Name:(NSString*)trackName Album:(NSString*)albumName Artist:(NSString*) artistName;
@end


@interface MusicBox : NSObject
@property (nonatomic,strong) NSString* DeviceName;
@property (nonatomic,strong) NSString* ID;
@property (nonatomic,strong) NSString* Theme;
@property (nonatomic,strong) NSString* User;

@property BOOL playing;
@property (nonatomic,strong) NSMutableArray* tracks;

+ (instancetype) musicBoxWithDictionary:(NSDictionary*) dict;


@end
