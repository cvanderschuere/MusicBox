//
//  MusicBox.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 6/29/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "MusicBox.h"

@implementation MusicBoxTrack
+(instancetype) trackWithService:(NSString*)serviceName Url:(NSURL*)url Name:(NSString *)trackName Artist:(NSString *)artistName{
    MusicBoxTrack *newTrack = [[MusicBoxTrack alloc] init];
    newTrack.service = serviceName;
    newTrack.url = url;
    newTrack.trackName = trackName;
    newTrack.artistName = artistName;
    
    return newTrack;
}

@end

@implementation MusicBox

+ (instancetype) musicBoxWithName:(NSString*) name{
    MusicBox *box = [[MusicBox alloc] init];
    box.title = name;
    box.tracks = [NSMutableArray array];
    box.links = [NSMutableArray array];
    box.playing = false;
    return box;
}


-(void)setTracksWithLinks:(NSMutableArray *)linkArray{
    self.links = linkArray;
    
    [self.tracks removeAllObjects];
    self.loaded = NO;
    for(MusicBoxTrack *linkedTrack in self.links){
        [SPTrack trackForTrackURL:linkedTrack.url inSession:[SPSession sharedSession] callback:^(SPTrack *track) {
            [self.tracks addObject:track];
            if (self.tracks.count == linkArray.count){
                //Load all tracks
                [SPAsyncLoading waitUntilLoaded:self.tracks timeout:10 then:^(NSArray *loadedItems, NSArray *notLoadedItems) {
                    //All tracks are loaded...load albums for artwork
                    NSMutableArray* imageArray = [NSMutableArray arrayWithCapacity:self.tracks.count];
                    for (SPTrack *track in loadedItems) {
                        [track.album.cover startLoading];
                        [imageArray addObject:track.album.cover];
                    }
                    
                    //Load all albums
                    [SPAsyncLoading waitUntilLoaded:imageArray timeout:10 then:^(NSArray *loadedItems, NSArray *notLoadedItems) {
                        NSLog(@"Loaded: %d Not-Loaded:%d",loadedItems.count,notLoadedItems.count);
                        self.loaded = YES;
                    }];
                }];
            }
        }];
    }
}

- (void) addTrackWithLink:(MusicBoxTrack*)link atIndex:(NSUInteger)idx{
    [self.links insertObject:link atIndex:idx];
    
    self.loaded = NO;
    [SPTrack trackForTrackURL:link.url inSession:[SPSession sharedSession] callback:^(SPTrack *track) {
        [self.tracks insertObject:track atIndex:idx];
        
        [SPAsyncLoading waitUntilLoaded:track timeout:10 then:^(NSArray *loadedItems, NSArray *notLoadedItems) {
            if (loadedItems.count>0) {
                SPTrack *track = loadedItems.lastObject;
                
                [SPAsyncLoading waitUntilLoaded:track.album.cover timeout:10 then:^(NSArray *loadedItems, NSArray *notLoadedItems) {
                    self.loaded = YES;
                }];
            }
        }];

    }];
    
}
- (void) removeTrackWithLink:(MusicBoxTrack*)link{
    NSUInteger idx = [self.links indexOfObject:link];
    [self.tracks removeObjectAtIndex:idx];
    [self.links removeObjectAtIndex:idx];
}

@end
