//
//  MusicBox.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 6/29/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "MusicBox.h"
#define LastFMAPIKey @"600be92e4856b530ec9ffaef2906e5a6"


@implementation MusicBoxTrack
+(instancetype) trackWithDict:(NSDictionary *)dict{
    NSLog(@"New Track Dict:%@",dict);
    MusicBoxTrack *newTrack = [[MusicBoxTrack alloc] init];
    newTrack.trackName = dict[@"Title"];
    newTrack.artistName = dict[@"ArtistName"];
    newTrack.albumName = dict[@"AlbumName"];
    
    newTrack.url = dict[@"ProviderID"];
    
    //Artwork
    NSString* urlString = dict[@"ArtworkURL"];
    if (![urlString isEqualToString:@""]) {
        newTrack.artworkURL = [NSURL URLWithString:urlString];
    }

 
    return newTrack;
}

- (BOOL) isEqual:(id)object{
    if ([object isKindOfClass:[MusicBoxTrack class]]) {
        MusicBoxTrack *compareTrack = (MusicBoxTrack*) object;
        return [self.url isEqualToString:compareTrack.url];
    }
    else
        return [super isEqual:object];
}

@end

@implementation MusicBox
+ (instancetype) musicBoxWithDictionary:(NSDictionary*) dict{
    MusicBox *box = [[MusicBox alloc] init];
    box.User = dict[@"User"];
    box.DeviceName = dict[@"DeviceName"];
    box.Theme = dict[@"Theme"];
    box.ID = dict[@"ID"];
    
    box.playing = [dict[@"Playing"] boolValue];
    
    box.tracks = [NSMutableArray array];
    
    return box;
}

@end
