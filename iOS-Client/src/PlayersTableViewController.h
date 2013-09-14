//
//  PlayersTableViewController.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/30/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "AppDelegate.h"
#import "MusicBox.h"

@interface PlayersTableViewController : UITableViewController <MDWampRpcDelegate>

@property (nonatomic,strong) MusicBox* selectedPlayer;
- (IBAction)cancelSelection:(id)sender;

@end
