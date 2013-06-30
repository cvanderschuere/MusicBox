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

+(instancetype) trackWithService:(NSString*)serviceName Url:(NSString*)url;
@end


@interface MusicBox : NSObject

@property (nonatomic,strong) NSString* title;

//track handling
@property (nonatomic, strong) NSMutableArray *tracks;
@property (nonatomic, strong) NSArray *queue;
@property BOOL loaded;

+ (instancetype) musicBoxWithName:(NSString*) name;

-(void) setTracksWithLinks:(NSArray*)linkArray;

@end
