//
//  MainViewController.m
//  PandoraController
//
//  Created by Chris Vanderschuere on 2/9/14.
//  Copyright (c) 2014 CDVConcepts. All rights reserved.
//

#import "MainViewController.h"
#import "AppDelegate.h"
#import "Station.h"
#import <AFNetworking/UIImageView+AFNetworking.h>
#import "Device.h"

@interface MainViewController ()

@end

@implementation MainViewController

- (void) setCurrentUser:(User *)currentUser{
    _currentUser = currentUser;
    
    //Do something now that we have a new user
    NSLog(@"User Added: %@",_currentUser);
    
    AppDelegate *delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    
    //Load user information
    [delegate.ws call:[NSString stringWithFormat:@"%@players",delegate.ws.baseURL] complete:^(NSString *callURI, id result, NSError *error) {
        //Get detailed information about each
        NSArray *ids = (NSArray*) result;

        [delegate.ws call:[NSString stringWithFormat:@"%@boxDetails",delegate.ws.baseURL] complete:^(NSString *callURI, id result, NSError *error) {
            NSMutableArray *deviceArray = [NSMutableArray array];
            
            //Add box for each
            for (NSString* key in result) {
                Device* newDevice = [Device deviceWithDict:[result[key] objectForKey:@"box"]];
                [deviceArray addObject:newDevice];
                
                //Subscribe to updates if allowed
                NSString* uri = [NSString stringWithFormat:@"%@%@/%@",delegate.ws.baseURL,self.currentUser.username,newDevice.identifier];
                NSDictionary* value = self.currentUser.pubSubPerms[uri];
                if (value != nil && value[@"CanSubscribe"]) {
                    //Subscribe to this update
                    NSLog(@"Subscribed: %@",uri);
                    [delegate.ws subscribeTopic:uri onEvent:^(id payload) {
                        [self handleEvent:payload withTopic:uri];
                    }];
                }
            }
            
            
            //Sort by device name
            [deviceArray sortUsingComparator:^NSComparisonResult(id obj1, id obj2) {
                return [[obj1 deviceName] compare:[obj2 deviceName]];
            }];
            
            _currentUser.devices = deviceArray;
            _currentUser.selectedDevice = deviceArray[0];
            [self updateWithDevice:deviceArray[0]];

            
            NSLog(@"Devices: %@",deviceArray);
            
        } args:ids,nil];
    } args:nil];
    
    [delegate.ws call:[NSString stringWithFormat:@"%@themes",delegate.ws.baseURL] complete:^(NSString *callURI, id result, NSError *error) {
        
        NSArray* resultArray = (NSArray*) result;
        
        NSMutableArray *stations = [NSMutableArray arrayWithCapacity:resultArray.count];
        for(NSDictionary* stationDict in resultArray){
            Station* station = [Station stationWithDictionary:stationDict];
            [stations addObject:station];
        }
        
        _currentUser.stations = stations;
        
        [[NSOperationQueue mainQueue] addOperationWithBlock:^{
            [self.stationCollectionView reloadData];
        }];
        
    } args:nil];
    
}

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
	// Do any additional setup after loading the view.
    
    self.navigationItem.hidesBackButton = YES;
    [self updateCurrentTrack:nil];
    
    self.blurView.tintColor = [UIColor blackColor];
    
    
    [[NSNotificationCenter defaultCenter] addObserverForName:@"Reauthenticated User" object:nil queue:[NSOperationQueue mainQueue] usingBlock:^(NSNotification *note) {
        User* reauthUser = (User*) note.object;
        [self setCurrentUser:reauthUser];
    }];
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - UICollectionView Delegate/Datasource
- (NSInteger) numberOfSectionsInCollectionView:(UICollectionView *)collectionView{
    return 1;
}
- (NSInteger) collectionView:(UICollectionView *)collectionView numberOfItemsInSection:(NSInteger)section{
    return self.currentUser.stations.count;
}
- (UICollectionViewCell*)collectionView:(UICollectionView *)collectionView cellForItemAtIndexPath:(NSIndexPath *)indexPath{
    UICollectionViewCell *cell = [collectionView dequeueReusableCellWithReuseIdentifier:@"StationCell" forIndexPath:indexPath];
    
    //Customize the cell for this indexPath
    Station *specificStation = (Station *) [self.currentUser.stations objectAtIndex:indexPath.item];
    
    UILabel *nameLabel = (UILabel*) [cell viewWithTag:200];
    UIImageView *artworkImageVew = (UIImageView*) [cell viewWithTag:100];
    
    nameLabel.text = specificStation.name;
    [artworkImageVew setImageWithURL:[NSURL URLWithString:specificStation.artworkURL]];
    
    return cell;
}

- (void) collectionView:(UICollectionView *)collectionView didSelectItemAtIndexPath:(NSIndexPath *)indexPath{
    Station *selectedStation = (Station *) [self.currentUser.stations objectAtIndex:indexPath.item];

    //Publish notice of theme change
    AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [del.ws publish:@{
                      @"command":@"updateTheme",
                      @"data":@{
                                @"ThemeID":selectedStation.themeID,
                                @"Name":selectedStation.name,
                                @"ArtworkURL":selectedStation.artworkURL
                                }
                      }
            toTopic:[NSString stringWithFormat:@"%@%@/%@",del.ws.baseURL,self.currentUser.username,self.currentUser.selectedDevice.identifier] excludeMe:YES
     ];
    
}



#pragma mark - Actions

- (IBAction)devicesAction:(id)sender {
    UIActionSheet *deviceSelector = [[UIActionSheet alloc] initWithTitle:@"Select Device" delegate:self cancelButtonTitle:nil destructiveButtonTitle:nil otherButtonTitles:nil];
    
    for( Device *d in self.currentUser.devices)  {
        [deviceSelector addButtonWithTitle:d.deviceName];
    }
    
    [deviceSelector addButtonWithTitle:@"Cancel"];
    deviceSelector.cancelButtonIndex = self.currentUser.devices.count;
    
    [deviceSelector showInView:self.view];
}

- (void) actionSheet:(UIActionSheet *)actionSheet clickedButtonAtIndex:(NSInteger)buttonIndex{
    if (buttonIndex >= self.currentUser.devices.count) {
        //Clicked cancel
        return;
    }
    
    
    self.currentUser.selectedDevice = self.currentUser.devices[buttonIndex];
    [self updateWithDevice:self.currentUser.selectedDevice];
    
}

#pragma mark - Events
- (void) handleEvent:(NSDictionary *)event withTopic:(NSString *)topicURI{
    //Extract command && data
    NSString* command = [event objectForKey:@"command"];
    NSLog(@"Command: %@",command);
    
    //Switch between commands
    if ([command isEqualToString:@"startedTrack"]) {
        //Extract data
        NSDictionary* data = [event objectForKey:@"data"];
        NSString* deviceID = data[@"deviceID"];
        
        //Find device
        Device* thisDevice = nil;
        for (Device* device in self.currentUser.devices) {
            if ([device.identifier isEqualToString:deviceID]) {
                thisDevice = device;
                break;
            }
        }
        
        //Extract track
        Track* newTrack = [Track trackWithDict:data[@"track"]];
        
        [thisDevice.tracks addObject:newTrack];
        thisDevice.playState = [NSNumber numberWithInt:DEVICE_PLAYING];
        
        //Update if you need to
        if ([thisDevice isEqual:self.currentUser.selectedDevice]) {
            [self updateWithDevice:self.currentUser.selectedDevice];
        }
        
        
    }else if([command isEqualToString:@"playTrack"] || [command isEqualToString:@"pauseTrack"]){
        NSArray* pathComps = [[NSURL URLWithString:topicURI] pathComponents];
        
        NSString* deviceID = pathComps[2];
        
        //Find device
        Device* thisDevice = nil;
        for (Device* device in self.currentUser.devices) {
            if ([device.identifier isEqualToString:deviceID]) {
                thisDevice = device;
                break;
            }
        }
        
        
        thisDevice.playState = [NSNumber numberWithInt:[command isEqualToString:@"playTrack"]?DEVICE_PLAYING:DEVICE_PAUSED];
        
        if (thisDevice == self.currentUser.selectedDevice) {
            [self updateWithDevice:self.currentUser.selectedDevice];
        }
    }else if([command isEqualToString:@"boxConnected"] || [command isEqualToString:@"boxDisconnected"]){
        NSArray* pathComps = [[NSURL URLWithString:topicURI] pathComponents];
        
        NSString* deviceID = pathComps[2];
        
        //Find device
        Device* thisDevice = nil;
        for (Device* device in self.currentUser.devices) {
            if ([device.identifier isEqualToString:deviceID]) {
                thisDevice = device;
                break;
            }
        }
        
        thisDevice.playState = [NSNumber numberWithInt:[command isEqualToString:@"boxDisconnected"]?DEVICE_OFFLINE:DEVICE_PAUSED];
        if (thisDevice == self.currentUser.selectedDevice) {
            [self updateWithDevice:self.currentUser.selectedDevice];
        }
    }
    

}

- (void) playDevice{
    self.currentUser.selectedDevice.playState = @(DEVICE_PLAYING);
    
    //Send play
    AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [del.ws publish:@{
                      @"command":@"playTrack",
                      }
            toTopic:[NSString stringWithFormat:@"%@%@/%@",del.ws.baseURL,self.currentUser.username,self.currentUser.selectedDevice.identifier]
          excludeMe:YES];
    
    [self updateWithDevice:self.currentUser.selectedDevice];
}
- (void) pauseDevice{
    self.currentUser.selectedDevice.playState = @(DEVICE_PAUSED);
    
    //Send pause
    AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [del.ws publish:@{
                      @"command":@"pauseTrack",
                      }
            toTopic:[NSString stringWithFormat:@"%@%@/%@",del.ws.baseURL,self.currentUser.username,self.currentUser.selectedDevice.identifier]
          excludeMe:YES];

 
    [self updateWithDevice:self.currentUser.selectedDevice];
}

- (IBAction)nextTrack{
    //Send next
    AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [del.ws publish:@{
                      @"command":@"nextTrack",
                      }
            toTopic:[NSString stringWithFormat:@"%@%@/%@",del.ws.baseURL,self.currentUser.username,self.currentUser.selectedDevice.identifier]
          excludeMe:YES];
}

#pragma mark - View Layout
- (void) updateCurrentTrack:(Track*)track{
    if (track == nil) {
        self.currentTrackAlbumArtwork.image = nil;
        self.currentTrackArtistName.text = @"";
        self.currentTrackTrackTitle.text = @"";
        
        self.nextButton.alpha = 0;
    }else{
        [self.currentTrackAlbumArtwork setImageWithURL:track.artworkURL];
        self.currentTrackArtistName.text = track.artistName;
        self.currentTrackTrackTitle.text = track.trackName;
        
        self.nextButton.alpha = 1;
    }
}

- (void) updateWithDevice:(Device*)device{
    self.title = device.deviceName;

    
    switch (device.playState.intValue) {
        case DEVICE_PLAYING:
        {
            UIBarButtonItem *item = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPause target:self action:@selector(pauseDevice)];
            item.tintColor = [UIColor blackColor];
            self.navigationItem.rightBarButtonItem = item;
            break;
        }
        case DEVICE_PAUSED:
        {
            UIBarButtonItem *item = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPlay target:self action:@selector(playDevice)];
            item.tintColor = [UIColor blackColor];
            self.navigationItem.rightBarButtonItem = item;
            break;
        }
        default:
        {
            UIBarButtonItem *item = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemStop target:self action:@selector(playDevice)];
            item.enabled = NO;
            item.tintColor = [UIColor blackColor];
            self.navigationItem.rightBarButtonItem = item;
            break;
        }
    }
    
    
    
    [self updateCurrentTrack:device.tracks.lastObject];
}

@end
