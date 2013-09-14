//
//  PlaylistViewController.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/31/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "PlaylistViewController.h"
#import "PlayersTableViewController.h"
#import "TrackSearchViewController.h"
#import "TrackCell.h"

@interface PlaylistViewController ()

@end

@implementation PlaylistViewController

- (void) setCurrentPlayer:(MusicBox *)currentPlayer{
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    
    //Cleanup from previous
    if (_currentPlayer) {        
        [delegate.websocketRequestQueue addOperationWithBlock:^(){
            //Unsubscribe to updates
            [delegate.ws unsubscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,_currentPlayer.User,_currentPlayer.ID]];
            
            //Subscribe to new topic
            [delegate.ws subscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,_currentPlayer.User,currentPlayer.ID] withDelegate:self];
            
            // TODO: Add queue request
            
        }];
    }
    else{
        //Just subscribe
        [delegate.websocketRequestQueue addOperationWithBlock:^(){
            //Subscribe to new topic
            [delegate.ws subscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,_currentPlayer.User,currentPlayer.ID] withDelegate:self];
            
            // TODO: Add queue request
            
        }];
    }
    
    _currentPlayer = currentPlayer;
    
    //Update top bottom
    if (_currentPlayer)
        self.playerButton.title = _currentPlayer.DeviceName;
    else
        self.playerButton.title = @"Select Player";
        
    //Save for later
    [[NSUserDefaults standardUserDefaults] setValue:_currentPlayer.ID forKey:@"previousPlayer"];
    [[NSUserDefaults standardUserDefaults]synchronize];
    
    //Refresh screen
    [self.collectionView reloadData];
}

- (void)viewDidLoad
{
    [super viewDidLoad];
	// Do any additional setup after loading the view.
    
    [self.playPauseButton setTitle:@"Play" forState:UIControlStateNormal];
    
    //Add refresh control
    self.refreshControl = [[UIRefreshControl alloc] init];
    self.refreshControl.tintColor = [UIColor lightGrayColor];
    [self.refreshControl addTarget:self action:@selector(refreshPlaylist:) forControlEvents:UIControlEventValueChanged];
    [self.collectionView addSubview:self.refreshControl];
    self.collectionView.alwaysBounceVertical = YES;
    
    [self.refreshControl beginRefreshing];
    
    //FIXME Refresh UI for previously selected player
    /*
    NSString* previousPlayerTitle = [[NSUserDefaults standardUserDefaults] valueForKey:@"previousPlayer"];
    if (previousPlayerTitle) {
        //Create player
        self.currentPlayer = [MusicBox musicBoxWithName:previousPlayerTitle];
    }
     */
}
- (void) observeValueForKeyPath:(NSString *)keyPath ofObject:(id)object change:(NSDictionary *)change context:(void *)context{
    if ([keyPath isEqualToString:@"artworkURL"]) {
        NSLog(@"Artwork Loaded");
    
        //Refresh playlist
        [self.collectionView reloadData];        
    }
}
- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark WAMP event 
- (void) onEvent:(NSString *)topicUri eventObject:(id)object{
    NSLog(@"Recieved Event:%@",object);
    
    [self.refreshControl endRefreshing];
    
    //TOD): Check if still same player ![self.currentPlayer.title isEqualToString:topicUri.lastPathComponent]
    if (![object isKindOfClass:[NSDictionary class]] ) {
        //Incorrect object recieved
        return;
    }
    
    //Form: baseURL+"/"+deviceName+"/currentQueue"
    //Follow api as define in apiary.io blueprint
    
    if ([[object objectForKey:@"command"] isEqualToString:@"statusUpdate"]) {
        NSDictionary *data = [object objectForKey:@"data"];
        
        //Play/Pause
        //FIXME self.currentPlayer.playing = [[data objectForKey:@"isPlaying"] boolValue];
        [self.playPauseButton setTitle:self.currentPlayer.playing?@"Pause":@"Play" forState:UIControlStateNormal];
        
        //Queue: merge
        if (![[data objectForKey:@"queue"] isKindOfClass:[NSNull class]]) {
            NSArray *queue = (NSArray*) [data objectForKey:@"queue"];
            NSMutableArray *recievedArray = [NSMutableArray arrayWithCapacity:queue.count];

            if (queue.count>0) {
                for(NSDictionary *track in queue){
                    MusicBoxTrack * addedTrack = [MusicBoxTrack trackWithDict:track];
                    
                    NSUInteger index = [self.currentPlayer.tracks indexOfObject:addedTrack];
                    if (index == NSNotFound) {
                        //Add new track
                        [recievedArray addObject:addedTrack];
                        [addedTrack addObserver:self forKeyPath:@"artworkURL" options:NSKeyValueObservingOptionNew context:NULL];
                    }
                    else{
                        //Add existing version
                        [recievedArray addObject:[self.currentPlayer.tracks objectAtIndex:index]];
                    }
                }
                self.currentPlayer.tracks = recievedArray;
                [self.collectionView reloadData];
            }
        }
    }
    else if ([[object objectForKey:@"command"] isEqualToString:@"addTrack"]){
        NSArray *tracks = [object objectForKey:@"data"];
        for (NSDictionary* track in tracks) {
            MusicBoxTrack * addedTrack = [MusicBoxTrack trackWithDict:track];
            
            [self.currentPlayer.tracks addObject:addedTrack];
            [addedTrack addObserver:self forKeyPath:@"artworkURL" options:NSKeyValueObservingOptionNew context:NULL];
        }
        [self.collectionView reloadData];
        
    }
    else if ([[object objectForKey:@"command"] isEqualToString:@"playTrack"]){
        [self.playPauseButton setTitle:@"Pause" forState:UIControlStateNormal];
        
    }
    else if ([[object objectForKey:@"command"] isEqualToString:@"pauseTrack"]){
        [self.playPauseButton setTitle:@"Play" forState:UIControlStateNormal];
    }
    else if ([[object objectForKey:@"command"] isEqualToString:@"startedTrack"]){
        MusicBoxTrack* track =  [MusicBoxTrack trackWithDict:[object[@"data"] objectForKey:@"track"]];
        [self.currentPlayer.tracks addObject:track];
        [self.collectionView reloadData];
    }
    
}

#pragma mark RPC Response
- (void) onResult:(id)result forCalledUri:(NSString *)callUri{
    if ([callUri isEqualToString:[NSString stringWithFormat:@"%@%@",baseURL,@"trackHistory"]]) {
        [self.refreshControl endRefreshing];
        NSLog(@"Result:%@",result);
    }

}
- (void) onError:(NSString *)errorUri description:(NSString *)errorDesc forCalledUri:(NSString *)callUri{
    NSLog(@"Error on RPC: %@",errorUri);
    [self.refreshControl endRefreshing];
}

-(void) refreshPlaylist:(UIRefreshControl*)sender{    
    //Refresh tracks of previous player
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [delegate.websocketRequestQueue addOperationWithBlock:^(){
        // TODO: Add queue request
        [delegate.ws call:[NSString stringWithFormat:@"%@%@",baseURL,@"trackHistory"] withDelegate:self args:self.currentPlayer.ID, nil];
    }];
}
#pragma mark - UICollectionView Datasource methods
-(NSInteger) collectionView:(UICollectionView *)collectionView numberOfItemsInSection:(NSInteger)section{
    return self.currentPlayer.tracks.count;
}
- (UICollectionViewCell*) collectionView:(UICollectionView *)collectionView cellForItemAtIndexPath:(NSIndexPath *)indexPath{
    TrackCell* cell = (TrackCell*)[collectionView dequeueReusableCellWithReuseIdentifier:@"trackCell" forIndexPath:indexPath];
    
    MusicBoxTrack* track = [self.currentPlayer.tracks objectAtIndex:indexPath.row];
    cell.trackTitle.text = track.trackName;
    
    if (track.artworkURL) {
        //Set by url
        [cell.albumArt setImageWithURL:track.artworkURL placeholderImage:[UIImage imageNamed:@"music-note.jpg"]];
    }
    else{
        //Use default while waiting
        cell.albumArt.image = [UIImage imageNamed:@"music-note.jpg"];
    }
    cell.albumArt.layer.cornerRadius = 5.0f;
    cell.albumArt.layer.masksToBounds = YES;
    
    return cell;
}
/*
#pragma mark Rotation Events
-(void)willRotateToInterfaceOrientation:(UIInterfaceOrientation)toInterfaceOrientation
                               duration:(NSTimeInterval)duration{
    
    //Get current layout
    UICollectionViewFlowLayout* layout = (UICollectionViewFlowLayout*) self.collectionView.collectionViewLayout;
    
    if (UIDeviceOrientationIsPortrait(toInterfaceOrientation)) {
        layout.scrollDirection = UICollectionViewScrollDirectionVertical;
        [self.collectionView reloadData];
    } else {
        layout.scrollDirection = UICollectionViewScrollDirectionHorizontal;
        [self.collectionView reloadData];
    }
}
*/

#pragma mark - UIStoryboard Segue Methods
- (IBAction)nextPressed:(id)sender {
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    NSString* username = @"christopher.vanderschuere@gmail.com";
    [delegate.websocketRequestQueue addOperationWithBlock:^{
        [delegate.ws publish:@{@"command": @"nextTrack"} toTopic:[NSString stringWithFormat:@"%@%@",baseURL,self.currentPlayer.ID] excludeMe:YES];
    }];
    
    //Animate deletion (only if not only song left)
    if (self.currentPlayer.tracks.count > 1) {
        [self.collectionView performBatchUpdates:^{
            [self.currentPlayer.tracks removeObjectAtIndex:0];
            [self.collectionView deleteItemsAtIndexPaths:@[[NSIndexPath indexPathForItem:0 inSection:0]]];
        } completion:^(BOOL finished) {
            
        }];
    }
}

- (IBAction)playPausePressed:(id)sender {
    NSString* username = @"christopher.vanderschuere@gmail.com";
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;

    if ([self.playPauseButton.titleLabel.text isEqualToString:@"Play"]) {
        [delegate.websocketRequestQueue addOperationWithBlock:^{
            [delegate.ws publish:@{@"command": @"playTrack"} toTopic:[NSString stringWithFormat:@"%@%@",baseURL,self.currentPlayer.ID] excludeMe:YES];
        }];
        [self.playPauseButton setTitle:@"Pause" forState:UIControlStateNormal];
    }
    else{
        [delegate.websocketRequestQueue addOperationWithBlock:^{
            [delegate.ws publish:@{@"command": @"pauseTrack"} toTopic:[NSString stringWithFormat:@"%@%@",baseURL,self.currentPlayer.ID] excludeMe:YES];
        }];
        [self.playPauseButton setTitle:@"Play" forState:UIControlStateNormal];
    }
}

-(IBAction)unwindFromPlayerSelection:(UIStoryboardSegue*)sender{
    //Set current Player base upon selected player...could have been done with a delegate
    MusicBox *selectedBox = [sender.sourceViewController selectedPlayer]; //Only title is passed right now
    
    if (selectedBox && ![selectedBox.ID isEqualToString:self.currentPlayer.ID]) {
        self.currentPlayer = selectedBox;
    }
    
}
-(IBAction)unwindFromTrackSelection:(UIStoryboardSegue *)sender{
    //Get selected track from 
    TrackSearchViewController* trackVC = sender.sourceViewController;
    NSLog(@"Track: %@",trackVC.selectedTrack.url);
    
    if (trackVC.selectedTrack && self.currentPlayer) {
        //Add locally
        [self.currentPlayer.tracks addObject:trackVC.selectedTrack];
        [trackVC.selectedTrack addObserver:self forKeyPath:@"artworkURL" options:NSKeyValueObservingOptionNew context:NULL];
        [self.collectionView reloadData];
        
        //Add track
        AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
        NSString* username = @"christopher.vanderschuere@gmail.com";
        
        [delegate.websocketRequestQueue addOperationWithBlock:^{
            //Create message
            NSDictionary *addedTrackMessage = @{@"command": @"addTrack",
                                                @"data":@[[trackVC.selectedTrack dictionaryWithValuesForKeys:@[@"trackName",@"artistName",@"albumName",@"url",@"service"]]]
                                                };
            
            [delegate.ws publish:addedTrackMessage toTopic:[NSString stringWithFormat:@"%@%@",baseURL,self.currentPlayer.ID] excludeMe:YES];
        }];
    }
    
}
@end
