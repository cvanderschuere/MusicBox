//
//  LiveSearch.h
//  Viva
//
//  Created by Daniel Kennett on 6/9/11.
//  For license information, see LICENSE.markdown
//

#import <Foundation/Foundation.h>

@interface LiveSearch : NSObject

-(id)initWithInitialSearch:(SPSearch *)aSearch;
-(void)clear;

@property (nonatomic, readwrite, strong) SPSearch *latestSearch;

@property (nonatomic, readonly, copy) NSArray *topTracks;
@property (nonatomic, readonly, copy) NSArray *topArtists;
@property (nonatomic, readonly, copy) NSArray *topAlbums;

@property (nonatomic, readonly, strong) id topHit;

@end
