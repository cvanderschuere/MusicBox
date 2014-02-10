//
//  MBWebsocket.m
//  MusicBox-Controller
//
//  Created by Chris Vanderschuere on 1/31/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import "MBWebsocket.h"

@interface MBWebsocket ()

@property (nonatomic,strong) NSString* username;
@property (nonatomic,strong) NSString* sessionID;

@end


@implementation MBWebsocket

- (instancetype) initWithURLRequest:(NSURLRequest *)server delegate:(id<MDWampClientDelegate>)delegate{
    
    self = [super initWithURLRequest:server delegate:delegate];
    
    if (self) {
        //Create Request Queue for websocket
        self.requestQueue = [[NSOperationQueue alloc] init];
        [self.requestQueue setSuspended:YES];
        
        [self setOnConnectionOpen:^(MDWamp *wsMD) {
            MBWebsocket *ws = (MBWebsocket*)wsMD;
            
            if (ws.username != nil && ws.sessionID != nil) {
                //Auth websocket
                //Authenticate websocket
                [ws authWithKey:ws.username Secret:ws.sessionID Extra:nil Success:^(NSString *answer) {
                    [ws.requestQueue setSuspended:NO];
                    
                    NSLog(@"Authenticated (reconnect)");
                    
                    //Assume user doesn't need to be updated
                    
                    NSDictionary *responseDict = (NSDictionary *)answer;

                    User* newUser = [[User alloc] init];
                    newUser.username = ws.username;
                    newUser.sessionID = ws.sessionID;
                    
                    //Format permissions
                    newUser.pubSubPerms = responseDict[@"PubSub"];
                    newUser.rpcPerms = responseDict[@"RPC"];

                    [[NSNotificationCenter defaultCenter] postNotificationName:@"Reauthenticated User" object:newUser];
                    
                } Error:^(NSString *procCall, NSString *errorURI, NSString *errorDetails) {
                    NSLog(@"Auth Error (reconnect): %@",errorDetails);
                }];
            }

        }];
        
        [self setOnConnectionClose:^(MDWamp *wsMD, NSInteger code, NSString *error) {
            MBWebsocket *ws = (MBWebsocket*)wsMD;
            
            [ws.requestQueue setSuspended:YES];
        }];
    }

    return self;
}

#pragma mark - Queued Methods
- (void)publish:(id)event toTopic:(NSString *)topicUri
{
	[self publish:event toTopic:topicUri excludeMe:NO];
}
- (void) publish:(id)payload toTopic:(NSString *)topicUri excludeMe:(BOOL)excludeMe{
    //Queue request
    [self.requestQueue addOperationWithBlock:^{
        [super publish:payload toTopic:topicUri excludeMe:excludeMe];
    }];
    
}

- (void) subscribeTopic:(NSString *)topicUri onEvent:(void (^)(id))eventBlock{
    [self.requestQueue addOperationWithBlock:^{
        [super subscribeTopic:topicUri onEvent:eventBlock];
    }];
}

/*
- (NSString *) call:(NSString *)procUri complete:(void (^)(NSString *, id, NSError *))completeBlock args:(id)firstArg, ...{
    va_list ap;
    va_start(ap, firstArg);
    NSString *result = [super call:[NSString stringWithFormat:@"%@%@",self.baseURL,procUri] complete:completeBlock args:ap];
    va_end(ap);
    
    return result;
}
*/

//Override to auth
- (void) reconnect
{
	if (![self isConnected]) {
        [self.requestQueue setSuspended:YES];

		[self disconnect];
		[self connect];
        
    }
}

#pragma mark - Custom Auth
- (void)authenticateWebsocketWithUsername:(NSString*)username Password:(NSString*)password Callback:(void(^)(User* user, NSError* error))callback{
    //Lookup sessionid for username/password
    [self call:[NSString stringWithFormat:@"%@user/startSession",self.baseURL] complete:^(NSString *callURI, id result, NSError *error) {
        //Check for error
        if (error != nil) {
            return callback(nil,[NSError errorWithDomain:@"Registration Error" code:500 userInfo:nil]);
        }

        NSDictionary* resultDict = (NSDictionary*) result;
        NSString *sessionID = resultDict[@"sessionID"];
        
        //Save for reconnect
        self.username = username;
        self.sessionID = sessionID;
        
        //Authenticate websocket
        [self authWithKey:username Secret:sessionID Extra:nil Success:^(NSString *answer) {
            NSDictionary *responseDict = (NSDictionary *)answer;
            
            NSLog(@"Authenticated");
            
            User* newUser = [[User alloc] init];
            newUser.username = username;
            newUser.sessionID = sessionID;
            
            //Format permissions
            newUser.pubSubPerms = responseDict[@"PubSub"];
            newUser.rpcPerms = responseDict[@"RPC"];
            
            callback(newUser,nil);
        } Error:^(NSString *procCall, NSString *errorURI, NSString *errorDetails) {
            NSLog(@"Auth Error: %@",errorDetails);
            return callback(nil,[NSError errorWithDomain:@"Authentication Error" code:500 userInfo:nil]);
        }];
        
    } args:@{
             @"username": username,
             @"password": password
             }, nil
     ];
}


@end
