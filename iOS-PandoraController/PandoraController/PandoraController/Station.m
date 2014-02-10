//
//  Station.m
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import "Station.h"

@implementation Station

//Keys in dictionary: ThemeID, Name, ArtworkURL, Type
+ (instancetype) stationWithDictionary:(NSDictionary *)dict{
    Station *newStation = [[Station alloc] init];
    
    //Set properties
    newStation.themeID = dict[@"ThemeID"];
    newStation.name = dict[@"Name"];
    newStation.artworkURL = dict[@"ArtworkURL"];
    
    return newStation;
}



@end
