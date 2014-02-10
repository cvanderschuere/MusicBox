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
#import <UIImageView+AFNetworking.h>
#import "SettingsViewController.h"

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
    self.infoButton.enabled = NO;
    
	// Do any additional setup after loading the view.
    self.devices = [NSMutableArray array];
    
    //Load user information
    [self.ws call:[NSString stringWithFormat:@"%@players",baseURL] success:^(NSString *callURI, id result) {
        //Get detailed information about each
        NSArray *ids = (NSArray*) result;
        
        [self.ws call:[NSString stringWithFormat:@"%@boxDetails",baseURL] success:^(NSString *callURI, id result) {
            //Add box for each
            for (NSString* key in result) {
                Device* newDevice = [Device deviceWithDict:[result[key] objectForKey:@"box"]];
                [self.devices addObject:newDevice];
                
                //Subscribe to updates if allowed
                NSString* uri = [NSString stringWithFormat:@"%@%@/%@",baseURL,self.currentUser.username,newDevice.identifier];
                NSDictionary* value = self.currentUser.pubSubPerms[uri];
                if (value != nil && value[@"CanSubscribe"]) {
                    //Subscribe to this update
                    NSLog(@"Subscribed: %@",uri);
                    [self.ws subscribeTopic:uri withDelegate:self];
                }
                
            }
            
            //Sort by device name
            [self.devices sortUsingComparator:^NSComparisonResult(id obj1, id obj2) {
                return [[obj1 deviceName] compare:[obj2 deviceName]];
            }];
            
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
#pragma mark - Segue
- (void) prepareForSegue:(UIStoryboardSegue *)segue sender:(id)sender{
    if ([segue.identifier isEqualToString:@"SettingsSegue"]) {
        SettingsViewController* settingsVC = (SettingsViewController*) [segue.destinationViewController topViewController];
        settingsVC.selectedDevice = self.selectedDevice;
    }
    
}

#pragma mark - UICollectionView Delegate
- (void) collectionView:(UICollectionView *)collectionView didSelectItemAtIndexPath:(NSIndexPath *)indexPath{
    if (collectionView == self.deviceCollectionView) {
        
        self.infoButton.enabled = YES;
        
        //Look up tracks for this device
        self.selectedDevice = self.devices[indexPath.row];
        self.title = self.selectedDevice.deviceName;
        
        [self.ws call:[NSString stringWithFormat:@"%@trackHistory",baseURL] success:^(NSString *callURI, id result) {
            NSMutableArray *tracks = [NSMutableArray array];
            for (NSDictionary* dict in result) {
                [tracks addObject:[Track trackWithDict:dict]];
            }
            
            self.selectedDevice.tracks = [[tracks reverseObjectEnumerator] allObjects].mutableCopy;
            [self.trackCollectionView reloadData];
            
        } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
            NSLog(@"Error getting tracks:%@",errorDescription);
        } args:@[self.selectedDevice.identifier,@15], nil];
        
        /*
        [self.ws call:[NSString stringWithFormat:@"%@recommendSongs",baseURL] success:^(NSString *callURI, id result) {
            for (NSDictionary* track in result) {
                NSLog(@"Track:%@",track);
                NSDictionary* startedTrack = @{@"command": @"startedTrack",
                                               @"data":@{@"track": track,
                                                         @"deviceID":self.selectedDevice.identifier
                                                         }
                                               };
                
                [self.ws publish:startedTrack toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,self.currentUser.username,self.selectedDevice.identifier]];
            }
        
        } error:^(NSString *callURI, NSString *errorURI, NSString *errorDescription) {
            
        } args:self.selectedDevice.identifier, nil];
        */
        
    }else if(collectionView == self.trackCollectionView){
        
        //Do something
  
        
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
        if (track.artworkURL) {
            [cell.artworkImageView setImageWithURL:track.artworkURL placeholderImage:[UIImage imageNamed:@"music-note.jpg"]];
        }
        cell.artworkImageView.layer.cornerRadius = 6.0f;
        
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
    }else{
        return nil;
    }
}

#pragma mark - MDWampEventDelegate

- (void) onEvent:(NSString *)topicUri eventObject:(id)object{
    NSLog(@"Object: %@",object);
    
    if ([object isKindOfClass:[NSDictionary class]]) {
        //Extract command && data
        NSString* command = [object objectForKey:@"command"];
        
        //Switch between commands
        if ([command isEqualToString:@"startedTrack"]) {
            //Extract data
            NSDictionary* data = [object objectForKey:@"data"];
            NSString* deviceID = data[@"deviceID"];
            
            //Find device
            Device* thisDevice = nil;
            for (Device* device in self.devices) {
                if ([device.identifier isEqualToString:deviceID]) {
                    thisDevice = device;
                    break;
                }
            }
            
            //Extract track
            Track* newTrack = [Track trackWithDict:data[@"track"]];
            
            [thisDevice.tracks insertObject:newTrack atIndex:0];
            thisDevice.playState = [NSNumber numberWithInt:DEVICE_PLAYING];
            
            
            //Update if you need to
            if ([thisDevice isEqual:self.selectedDevice]) {
                [self.trackCollectionView insertItemsAtIndexPaths:@[[NSIndexPath indexPathForItem:0 inSection:0]]];
            }
            
            
        }else if([command isEqualToString:@"playTrack"] || [command isEqualToString:@"pauseTrack"]){
            NSArray* pathComps = [[NSURL URLWithString:topicUri] pathComponents];
            
            NSString* deviceID = pathComps[2];
            
            //Find device
            Device* thisDevice = nil;
            for (Device* device in self.devices) {
                if ([device.identifier isEqualToString:deviceID]) {
                    thisDevice = device;
                    break;
                }
            }
            
            
            thisDevice.playState = [NSNumber numberWithInt:[command isEqualToString:@"playTrack"]?DEVICE_PLAYING:DEVICE_PAUSED];
        }else if([command isEqualToString:@"boxConnected"] || [command isEqualToString:@"boxDisconnected"]){
            NSArray* pathComps = [[NSURL URLWithString:topicUri] pathComponents];
            
            NSString* deviceID = pathComps[2];
            
            //Find device
            Device* thisDevice = nil;
            for (Device* device in self.devices) {
                if ([device.identifier isEqualToString:deviceID]) {
                    thisDevice = device;
                    break;
                }
            }
            
            thisDevice.playState = [NSNumber numberWithInt:[command isEqualToString:@"boxDisconnected"]?DEVICE_OFFLINE:DEVICE_PAUSED];
        }
        

    }
}

@end
