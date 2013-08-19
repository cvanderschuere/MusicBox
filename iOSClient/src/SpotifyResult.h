//
//  SpotifyResult.h
//  MusicBox-iOS
//
//  Created by Chris Vanderschuere on 8/15/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface SpotifyResult : NSObject

@property (nonatomic) NSString* name;
@property (nonatomic) NSString* album;
@property (nonatomic) NSString* artist;
@property (nonatomic) NSString* href;

@end
