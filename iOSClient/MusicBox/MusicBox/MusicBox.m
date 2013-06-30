//
//  MusicBox.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 6/29/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "MusicBox.h"

@implementation MusicBoxTrack
+(instancetype) trackWithService:(NSString*)serviceName Url:(NSString*)url{
    MusicBoxTrack *newTrack = [[MusicBoxTrack alloc] init];
    newTrack.service = serviceName;
    newTrack.url = url;
    return newTrack;
}

@end

@implementation MusicBox

+ (instancetype) musicBoxWithName:(NSString*) name{
    MusicBox *box = [[MusicBox alloc] init];
    box.title = name;
    box.tracks = [NSMutableArray array];
    return box;
}


-(void)setTracksWithLinks:(NSArray *)linkArray{
    [self.tracks removeAllObjects];
    self.loaded = NO;
    for(MusicBoxTrack *linkedTrack in linkArray){
        [SPTrack trackForTrackURL:[NSURL URLWithString:linkedTrack.url] inSession:[SPSession sharedSession] callback:^(SPTrack *track) {
            [self.tracks addObject:track];
            if (self.tracks.count == linkArray.count)
                self.loaded = YES;
        }];
    }
}

@end
