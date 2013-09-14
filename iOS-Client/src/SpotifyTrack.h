//
//  SpotifyTrack.h
//  MusicBox-iOS
//
//  Created by Chris Vanderschuere on 8/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>
#import "SpotifyAlbum.h"
#import "SpotifyArtist.h"

@interface SpotifyTrack : NSObject

@property (nonatomic) NSString *name;
@property (nonatomic) NSString *spotifyID;

//Relationships
@property (nonatomic) SpotifyAlbum *album;
@property (nonatomic) NSArray* artists;
@end
