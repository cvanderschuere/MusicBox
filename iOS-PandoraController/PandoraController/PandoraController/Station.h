//
//  Station.h
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface Station : NSObject
@property (strong, nonatomic) NSString *themeID;
@property (strong, nonatomic) NSString *name;
@property (strong, nonatomic) NSString *artworkURL;


+(instancetype)stationWithDictionary:(NSDictionary*)dict;

@end
