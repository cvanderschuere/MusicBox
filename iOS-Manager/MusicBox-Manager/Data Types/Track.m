//
//  Track.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "Track.h"

@implementation Track

+ (instancetype) trackWithDict:(NSDictionary*)dict{
    Track* newTrack = [[Track alloc] init];
    newTrack.trackName = dict[@"Title"];
    newTrack.artistName = dict[@"ArtistName"];
    newTrack.albumName = dict[@"AlbumName"];
    newTrack.artworkURL = dict[@"ArtworkURL"];
    //newTrack.date = dict[@"ArtworkURL"]; //Convert string to date
    return newTrack;
}

@end
