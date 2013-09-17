//
//  MainViewController.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "MainViewController.h"
#import "AppDelegate.h" //for baseURL
#import "DeviceCell.h"
#import "TrackCell.h"

@interface MainViewController ()

@end

@implementation MainViewController

- (id)initWithNibName:(NSString *)nibNameOrNil bundle:(NSBundle *)nibBundleOrNil
{
    self = [super initWithNibName:nibNameOrNil bundle:nibBundleOrNil];
    if (self) {
        // Custom initialization
    }
    return self;
}


- (void)viewDidLoad
{
    [super viewDidLoad];
    
    self.title = nil;
    
	// Do any additional setup after loading the view.
    self.devices = [NSMutableArray array];
    
    //Load user information
    [self.ws call:[NSString stringWithFormat:@"%@players",baseURL] success:^(NSString *callURI, id result) {
        //Get detailed information about each
        NSArray *ids = (NSArray*) result;
        
        [self.ws call:[NSString stringWithFormat:@"%@boxDetails",baseURL] success:^(NSString *callURI, id result) {
            //Add box for each
            for (NSString* key in result) {
                [self.devices addObject:[Device deviceWithDict:result[key]]];
            }
            
            [self.deviceCollectionView reloadData];
            
        } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
            NSLog(@"Error:%@",errorDescription);
        } args:ids, nil];

        
    } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
        NSLog(@"Error:%@",errorDescription);
    } args:self.currentUser.username,nil];
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - UICollectionView Delegate
- (void) collectionView:(UICollectionView *)collectionView didSelectItemAtIndexPath:(NSIndexPath *)indexPath{
    if (collectionView == self.deviceCollectionView) {
        //Look up tracks for this device
        self.selectedDevice = self.devices[indexPath.row];
        self.title = self.selectedDevice.deviceName;
        
        [self.ws call:[NSString stringWithFormat:@"%@trackHistory",baseURL] success:^(NSString *callURI, id result) {
            NSMutableArray *tracks = [NSMutableArray array];
            for (NSDictionary* dict in result) {
                [tracks addObject:[Track trackWithDict:dict]];
            }
            self.selectedDevice.tracks = tracks;
            [self.trackCollectionView reloadData];
            
        } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
            NSLog(@"Error getting tracks:%@",errorDescription);
        } args:self.selectedDevice.identifier, nil];
        
    }else if(collectionView == self.trackCollectionView){
        
    }
}

#pragma mark - UICollection Datasource
- (NSInteger) numberOfSectionsInCollectionView:(UICollectionView *)collectionView{
    return 1;
}
- (NSInteger) collectionView:(UICollectionView *)collectionView numberOfItemsInSection:(NSInteger)section{
    if (collectionView == self.trackCollectionView) {
        return self.selectedDevice.tracks.count;
    }else if(collectionView == self.deviceCollectionView){
        return self.devices.count;
    }
    else{
        return 0;
    }
}
- (UICollectionViewCell*) collectionView:(UICollectionView *)collectionView cellForItemAtIndexPath:(NSIndexPath *)indexPath{
    if (collectionView == self.trackCollectionView) {
        TrackCell *cell = (TrackCell*) [collectionView dequeueReusableCellWithReuseIdentifier:@"trackCell" forIndexPath:indexPath];
        
        Track* track = self.selectedDevice.tracks[indexPath.row];
        cell.trackLabel.text = track.trackName;
        cell.artistLabel.text = track.artistName;
        
        return cell;

    }else if(collectionView == self.deviceCollectionView){
        DeviceCell *cell = (DeviceCell*)[collectionView dequeueReusableCellWithReuseIdentifier:@"deviceCell" forIndexPath:indexPath];
        cell.nameLabel.text = [self.devices[indexPath.row] deviceName];
        return cell;
    }
    else{
        return nil;
    }
}

- (UICollectionReusableView*) collectionView:(UICollectionView *)collectionView viewForSupplementaryElementOfKind:(NSString *)kind atIndexPath:(NSIndexPath *)indexPath{
    if (collectionView == self.trackCollectionView) {
        return [collectionView dequeueReusableSupplementaryViewOfKind:kind withReuseIdentifier:@"trackHeader" forIndexPath:indexPath];
    }
}

@end
