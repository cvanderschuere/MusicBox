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
+(instancetype) trackWithService:(NSString*)serviceName Url:(NSString*)url Name:(NSString *)trackName Album:(NSString *)albumName Artist:(NSString *)artistName{
    MusicBoxTrack *newTrack = [[MusicBoxTrack alloc] init];
    newTrack.service = serviceName;
    newTrack.url = url;
    newTrack.trackName = trackName;
    newTrack.artistName = artistName;
    newTrack.albumName = albumName;
    
    //Fetch Album artwork from last.fm (Do in background)
       
    //Encode as utf-8
    NSString *escapedArtist = (NSString *)CFBridgingRelease(CFURLCreateStringByAddingPercentEscapes(
                                                                                                    NULL,
                                                                                                    (__bridge CFStringRef) artistName,
                                                                                                    NULL,
                                                                                                    CFSTR("!*'();:@&=+$,/?%#[]"),
                                                                                                    kCFStringEncodingUTF8));
    NSString *escapedAlbum = (NSString *)CFBridgingRelease(CFURLCreateStringByAddingPercentEscapes(
                                                                                                    NULL,
                                                                                                    (__bridge CFStringRef) albumName,
                                                                                                    NULL,
                                                                                                    CFSTR("!*'();:@&=+$,/?%#[]"),
                                                                                                    kCFStringEncodingUTF8));
    
    //Load album artwork url
    NSString *requestString = [NSString stringWithFormat:@"http://ws.audioscrobbler.com/2.0/?method=album.getInfo&format=json&api_key=%@&artist=%@&album=%@",LastFMAPIKey,escapedArtist,escapedAlbum];

    
    //Make request
    NSURLRequest* request = [NSURLRequest requestWithURL:[NSURL URLWithString:requestString]];
   [NSURLConnection sendAsynchronousRequest:request queue:[NSOperationQueue mainQueue] completionHandler:^(NSURLResponse * response, NSData * data, NSError * error) {
       if (error) {
           NSLog(@"Error(Last.fm):%@",error);
       }
       else{
           //parse response
           NSDictionary* responseDict = [NSJSONSerialization JSONObjectWithData:data options:0 error:NULL];
           
           //Get large image dict
           NSDictionary *largeImageDict = [[responseDict[@"album"] objectForKey:@"image"] objectAtIndex:2];
           newTrack.artworkURL = [NSURL URLWithString:[largeImageDict objectForKey:@"#text"]];
       }
   }];
    
       
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

@end
