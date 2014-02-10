//
//  Track.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface Track : NSObject

@property (nonatomic,strong) NSString *trackName;
@property (nonatomic,strong) NSString *artistName;
@property (nonatomic,strong) NSString *albumName;

@property (nonatomic,strong) NSURL * artworkURL;
@property (nonatomic,strong) NSDate *date;


+ (instancetype) trackWithDict:(NSDictionary*)dict;

@end
