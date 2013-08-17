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
@property (nonatomic,strong) NSString * artistName;

+(instancetype) trackWithService:(NSString*)serviceName Url:(NSURL*)url Name:(NSString*)trackName Artist:(NSString*) artistName;
@end


@interface MusicBox : NSObject

@property (nonatomic,strong) NSString* title;
@property BOOL playing;

//track handling
@property (nonatomic, strong) NSMutableArray *tracks; //Sp_track instances
@property (nonatomic, strong) NSMutableArray *links; //Links in the form of MusicBoxTrack
@property BOOL loaded;

+ (instancetype) musicBoxWithName:(NSString*) name;

-(void) setTracksWithLinks:(NSArray*)linkArray;
- (void) addTrackWithLink:(MusicBoxTrack*)link atIndex:(NSUInteger)idx;
- (void) removeTrackWithLink:(MusicBoxTrack*)link;

@end
