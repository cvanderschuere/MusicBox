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
    NSString* username = @"christopher.vanderschuere@gmail.com";
    
    //Cleanup from previous
    if (_currentPlayer) {
        [_currentPlayer removeObserver:self forKeyPath:@"loaded"];
        
        [delegate.requestQueue addOperationWithBlock:^(){
            //Unsubscribe to updates
            [delegate.ws unsubscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,_currentPlayer.title]];
        }];
    }
    
    _currentPlayer = currentPlayer;
    
    //Add observer
    [_currentPlayer addObserver:self forKeyPath:@"loaded" options:NSKeyValueObservingOptionInitial context:NULL];
    
    [delegate.requestQueue addOperationWithBlock:^(){
        //Subscribe to updates
        [delegate.ws subscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] withDelegate:self];
    }];

    
    //Update top bottom
    if (_currentPlayer)
        self.playerButton.title = _currentPlayer.title;
    else
        self.playerButton.title = @"Select Player";
        
    //Save for later
    [[NSUserDefaults standardUserDefaults] setValue:_currentPlayer.title forKey:@"previousPlayer"];
    [[NSUserDefaults standardUserDefaults]synchronize];
}

- (void)viewDidLoad
{
    [super viewDidLoad];
	// Do any additional setup after loading the view.
    
    //Add refresh control
    UIRefreshControl *refreshControl = [[UIRefreshControl alloc] init];
    refreshControl.tintColor = [UIColor lightGrayColor];
    [refreshControl addTarget:self action:@selector(refreshPlaylist:) forControlEvents:UIControlEventValueChanged];
    [self.collectionView addSubview:refreshControl];
    self.collectionView.alwaysBounceVertical = YES;
    
    //Refresh UI for previously selected player
    NSString* previousPlayerTitle = [[NSUserDefaults standardUserDefaults] valueForKey:@"previousPlayer"];
    if (previousPlayerTitle) {
        //Create player
        self.currentPlayer = [MusicBox musicBoxWithName:previousPlayerTitle];
        
        NSString* username = @"christopher.vanderschuere@gmail.com";
        NSString* password = @"Example";
       
        AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
        
        [delegate.requestQueue addOperationWithBlock:^(){
            //Request tracks of previous player
            [delegate.ws call:[NSString stringWithFormat:@"%@currentQueueRequest",baseURL] withDelegate:self args:username,password,self.currentPlayer.title, nil];
        }];
    }
}
- (void) observeValueForKeyPath:(NSString *)keyPath ofObject:(id)object change:(NSDictionary *)change context:(void *)context{
    if ([keyPath isEqualToString:@"loaded"]) {
        NSLog(@"Loaded");
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
    NSLog(@"Recieved Event:%@",topicUri);
    NSString* username = @"christopher.vanderschuere@gmail.com";
    NSString* deviceName = @"LivingRoom";
    
    //Form: baseURL+username+"/"+deviceName+"/currentQueue"
    if ([topicUri isEqualToString:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,deviceName]]&& [object isKindOfClass:[NSArray class]]) {
        NSMutableArray *recievedTracks = [NSMutableArray arrayWithCapacity:[object count]];
        for(NSDictionary *item in object){
            [recievedTracks addObject:[MusicBoxTrack trackWithService:item[@"Service"] Url:item[@"URL"]]];
        }
        
        [self.currentPlayer setTracksWithLinks:recievedTracks];
    }
}

#pragma mark RPC Response
- (void) onResult:(id)result forCalledUri:(NSString *)callUri{
    //NOTHING TO DO YET
}
- (void) onError:(NSString *)errorUri description:(NSString *)errorDesc forCalledUri:(NSString *)callUri{
    NSLog(@"Error on RPC: %@",errorUri);
}

-(void) refreshPlaylist:(UIRefreshControl*)sender{    
    //Refresh tracks of previous player
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [delegate.requestQueue addOperationWithBlock:^(){
        [delegate.ws call:[NSString stringWithFormat:@"%@currentQueueRequest",baseURL] withDelegate:self args:@"christopher.vanderschuere@gmail.com",@"ExamplePassword",self.currentPlayer.title, nil];
    }];
    
    [sender endRefreshing];
}
#pragma mark - UICollectionView Datasource methods
-(NSInteger) collectionView:(UICollectionView *)collectionView numberOfItemsInSection:(NSInteger)section{
    return self.currentPlayer.tracks.count;
}
- (UICollectionViewCell*) collectionView:(UICollectionView *)collectionView cellForItemAtIndexPath:(NSIndexPath *)indexPath{
    TrackCell* cell = (TrackCell*)[collectionView dequeueReusableCellWithReuseIdentifier:@"trackCell" forIndexPath:indexPath];
    
    SPTrack* track = [self.currentPlayer.tracks objectAtIndex:indexPath.row];
    cell.trackTitle.text = track.name;
    
    //Load art in background
    cell.albumArt.image = track.album.cover.image;
    
    return cell;
}

#pragma mark - UIStoryboard Segue Methods
- (IBAction)nextPressed:(id)sender {
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    NSString* username = @"christopher.vanderschuere@gmail.com";
    [delegate.requestQueue addOperationWithBlock:^{
        [delegate.ws publish:@"NextTrack" toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
    }];
}

- (IBAction)playPausePressed:(id)sender {
}

-(IBAction)unwindFromPlayerSelection:(UIStoryboardSegue*)sender{
    //Set current Player base upon selected player...could have been done with a delegate
    MusicBox *selectedBox = [sender.sourceViewController selectedPlayer]; //Only title is passed right now
    
    if (selectedBox && ![selectedBox.title isEqualToString:self.currentPlayer.title]) {
        self.currentPlayer = selectedBox;
    }
    
}
-(IBAction)unwindFromTrackSelection:(UIStoryboardSegue *)sender{
    //Get selected track from 
    TrackSearchViewController* trackVC = sender.sourceViewController;
    NSLog(@"Track: %@",trackVC.selectedTrackURL);
    
    if (trackVC.selectedTrackURL && self.currentPlayer) {
        //Add track
        AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
        NSString* username = @"christopher.vanderschuere@gmail.com";
        NSString* serviceType = @"Spotify";
        [delegate.requestQueue addOperationWithBlock:^{
            [delegate.ws publish:[NSString stringWithFormat:@"AddTrack,%@,%@",serviceType,trackVC.selectedTrackURL]toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
        }];
    }
    
}
@end
