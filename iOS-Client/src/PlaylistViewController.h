//
//  PlaylistViewController.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/31/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "AppDelegate.h"
#import "MusicBox.h"

@interface PlaylistViewController : UIViewController <UICollectionViewDelegateFlowLayout,UICollectionViewDataSource,MDWampEventDelegate,MDWampRpcDelegate>

@property (nonatomic,strong) MusicBox *currentPlayer;
@property (weak, nonatomic) IBOutlet UIBarButtonItem *playerButton;
@property (weak, nonatomic) IBOutlet UICollectionView *collectionView;
@property (weak, nonatomic) IBOutlet UIButton *playPauseButton;
@property (weak, nonatomic) IBOutlet UIButton *nextButton;
@property (nonatomic,strong) UIRefreshControl *refreshControl;

- (IBAction)nextPressed:(id)sender;
- (IBAction)playPausePressed:(id)sender;
-(IBAction)unwindFromPlayerSelection:(UIStoryboardSegue*)sender;
-(IBAction)unwindFromTrackSelection:(UIStoryboardSegue*)sender;

@end
