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
@property (nonatomic,strong) NSURL *url;

@property (nonatomic,strong) NSString *trackName;
@property (nonatomic,strong) NSString *artistName;
@property (nonatomic,strong) NSString *albumName;

@property (nonatomic,strong) NSURL *artworkURL;

+(instancetype) trackWithService:(NSString*)serviceName Url:(NSURL*)url Name:(NSString*)trackName Album:(NSString*)albumName Artist:(NSString*) artistName;
@end


@interface MusicBox : NSObject

@property (nonatomic,strong) NSString* title;
@property BOOL playing;

//track handling
@property (nonatomic, strong) NSMutableArray *tracks; //MusicBoxTrack instances
@property (nonatomic, strong) NSMutableArray *links; //Links in the form of MusicBoxTrack
@property BOOL loaded;

+ (instancetype) musicBoxWithName:(NSString*) name;


@end
