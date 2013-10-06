//
//  MainViewController.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "User.h"
#import "Device.h"

@interface MainViewController : UIViewController <MDWampEventDelegate,UICollectionViewDataSource,UICollectionViewDelegate>

@property (nonatomic,strong) User *currentUser;
@property (nonatomic,strong) MDWamp *ws;

//User information
@property (nonatomic,strong) NSMutableArray *devices; //array of Device*
@property (nonatomic,strong) Device* selectedDevice;

@property (weak, nonatomic) IBOutlet UICollectionView *trackCollectionView;
@property (weak, nonatomic) IBOutlet UICollectionView *deviceCollectionView;


@end
