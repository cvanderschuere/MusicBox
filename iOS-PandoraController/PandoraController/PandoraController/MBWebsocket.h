//
//  MBWebsocket.h
//  MusicBox-Controller
//
//  Created by Chris Vanderschuere on 1/31/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import "User.h"

@interface MBWebsocket : MDWamp

@property (nonatomic,strong) NSString *baseURL;
@property (nonatomic, strong) NSOperationQueue *requestQueue;

- (void)authenticateWebsocketWithUsername:(NSString*)username Password:(NSString*)password Callback:(void(^)(User* user, NSError* error))callback;

@end
