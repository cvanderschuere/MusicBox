//
//  MainViewController.h
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "User.h"
#import <FXBlurView/FXBlurView.h>

@interface MainViewController : UIViewController <UICollectionViewDataSource,UICollectionViewDelegate,UIActionSheetDelegate>

@property (weak, nonatomic) IBOutlet UIImageView *currentTrackAlbumArtwork;
@property (weak, nonatomic) IBOutlet UILabel *currentTrackArtistName;
@property (weak, nonatomic) IBOutlet UILabel *currentTrackTrackTitle;
@property (weak, nonatomic) IBOutlet UICollectionView *stationCollectionView;

@property (weak, nonatomic) IBOutlet FXBlurView *blurView;

@property (strong, nonatomic) User* currentUser;

- (IBAction)devicesAction:(id)sender;
- (IBAction)nextTrack;

@property (weak, nonatomic) IBOutlet UIButton *nextButton;

@end
