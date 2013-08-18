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
    NSString* password = @"Example";

    
    //Cleanup from previous
    if (_currentPlayer) {        
        [delegate.websocketRequestQueue addOperationWithBlock:^(){
            //Unsubscribe to updates
            [delegate.ws unsubscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,_currentPlayer.title]];
            
            //Subscribe to new topic
            [delegate.ws subscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,currentPlayer.title] withDelegate:self];
            
            //Request tracks of previous player
            [delegate.ws call:[NSString stringWithFormat:@"%@currentQueueRequest",baseURL] withDelegate:self args:username,password,currentPlayer.title, nil];

        }];
    }
    else{
        //Just subscribe
        [delegate.websocketRequestQueue addOperationWithBlock:^(){
            //Subscribe to new topic
            [delegate.ws subscribeTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,currentPlayer.title] withDelegate:self];
            
            //Request tracks of previous player
            [delegate.ws call:[NSString stringWithFormat:@"%@currentQueueRequest",baseURL] withDelegate:self args:username,password,currentPlayer.title, nil];

        }];
    }
    
    _currentPlayer = currentPlayer;
    
    //Update top bottom
    if (_currentPlayer)
        self.playerButton.title = _currentPlayer.title;
    else
        self.playerButton.title = @"Select Player";
        
    //Save for later
    [[NSUserDefaults standardUserDefaults] setValue:_currentPlayer.title forKey:@"previousPlayer"];
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
    
    //Refresh UI for previously selected player
    NSString* previousPlayerTitle = [[NSUserDefaults standardUserDefaults] valueForKey:@"previousPlayer"];
    if (previousPlayerTitle) {
        //Create player
        self.currentPlayer = [MusicBox musicBoxWithName:previousPlayerTitle];
    }
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
    
    //Form: baseURL+username+"/"+deviceName+"/currentQueue"
    //Follow api as define in apiary.io blueprint
    
    if ([[object objectForKey:@"command"] isEqualToString:@"statusUpdate"]) {
        NSDictionary *data = [object objectForKey:@"data"];
        
        //Play/Pause
        self.currentPlayer.playing = [[data objectForKey:@"isPlaying"] boolValue];
        [self.playPauseButton setTitle:self.currentPlayer.playing?@"Pause":@"Play" forState:UIControlStateNormal];
        
        //Queue: merge
        if (![[data objectForKey:@"queue"] isKindOfClass:[NSNull class]]) {
            NSArray *queue = (NSArray*) [data objectForKey:@"queue"];
            NSMutableArray *recievedArray = [NSMutableArray arrayWithCapacity:queue.count];

            if (queue.count>0) {
                for(NSDictionary *track in queue){
                    MusicBoxTrack * addedTrack = [MusicBoxTrack trackWithService:track[@"Service"] Url:track[@"URL"] Name:track[@"TrackName"] Album:track[@"AlbumName"]Artist: track[@"ArtistName"]];
                    
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
            MusicBoxTrack * addedTrack = [MusicBoxTrack trackWithService:track[@"service"] Url:track[@"url"] Name:track[@"trackName"] Album:track[@"albumName"]Artist: track[@"artistName"]];
            
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
    
    /*
    //Queue update
    if ([topicUri isEqualToString:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,deviceName]]&& [object isKindOfClass:[NSArray class]]) {
        NSMutableArray *recievedTracks = [NSMutableArray arrayWithCapacity:[object count]];
        for(NSDictionary *item in object){
            //FIXME
            [recievedTracks addObject:[MusicBoxTrack trackWithService:item[@"Service"] Url:[NSURL URLWithString:item[@"URL"]]Name:@"Blank" Album:@"Album" Artist:@"Blank"]];
        }
        
        self.currentPlayer.tracks = recievedTracks;
    }
    //Status update
    else if ([topicUri isEqualToString:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,deviceName]]&& [object isKindOfClass:[NSDictionary class]]) {
        NSDictionary* statusObject = (NSDictionary*) object;
        
        //Play/Pause
        self.currentPlayer.playing = [[statusObject objectForKey:@"isPlaying"] boolValue];
        [self.playPauseButton setTitle:self.currentPlayer.playing?@"Pause":@"Play" forState:UIControlStateNormal];
        
        //Queue
        if (![[statusObject objectForKey:@"queue"] isKindOfClass:[NSNull class]]) {
            NSArray *queue = (NSArray*) [statusObject objectForKey:@"queue"];
            if (queue.count>0) {
                NSMutableArray *recievedTracks = [NSMutableArray arrayWithCapacity:[object count]];
                for(NSDictionary *item in queue){
                    // FIXME Album & artist name
                    [recievedTracks addObject:[MusicBoxTrack trackWithService:item[@"Service"] Url:[NSURL URLWithString:item[@"URL"]]Name:@"Blank" Album:@"Album" Artist:@"Unknown"]];
                }
                self.currentPlayer.tracks = recievedTracks;
            }
        }
    }
    else if ([object isKindOfClass:[NSString class]]){
        //Split string into components
        NSArray* components = [object componentsSeparatedByString:@","];
        if (components.count >0) {
            //Determine recieved command
            if ([components[0] isEqualToString:@"PlayTrack"]) {
                //Update display
                [self.playPauseButton setTitle:@"Pause" forState:UIControlStateNormal];
            }
            else if ([components[0] isEqualToString:@"PauseTrack"]) {
                //Update display
                [self.playPauseButton setTitle:@"Play" forState:UIControlStateNormal];
            }
            else if ([components[0] isEqualToString:@"AddTrack"]){
                if (components.count >=3 ) {
                    //FIXME artist & album name
                    MusicBoxTrack* addedTrack = [MusicBoxTrack trackWithService:components[1] Url:[NSURL URLWithString:components[2]]Name:@"Blank" Album:@"Album" Artist:@"Unknown"];
                    [self.currentPlayer.tracks addObject:addedTrack]; //Add to back
                    [addedTrack addObserver:self forKeyPath:@"artworkURL" options:NSKeyValueObservingOptionNew context:NULL];
                }
            }
        }
    }
    */
    
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
    [delegate.websocketRequestQueue addOperationWithBlock:^(){
        [delegate.ws call:[NSString stringWithFormat:@"%@currentQueueRequest",baseURL] withDelegate:self args:@"christopher.vanderschuere@gmail.com",@"ExamplePassword",self.currentPlayer.title, nil];
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


#pragma mark - UIStoryboard Segue Methods
- (IBAction)nextPressed:(id)sender {
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    NSString* username = @"christopher.vanderschuere@gmail.com";
    [delegate.websocketRequestQueue addOperationWithBlock:^{
        [delegate.ws publish:@{@"command": @"nextTrack"} toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
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
            [delegate.ws publish:@{@"command": @"playTrack"} toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
        }];
        [self.playPauseButton setTitle:@"Pause" forState:UIControlStateNormal];
    }
    else{
        [delegate.websocketRequestQueue addOperationWithBlock:^{
            [delegate.ws publish:@{@"command": @"pauseTrack"} toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
        }];
        [self.playPauseButton setTitle:@"Play" forState:UIControlStateNormal];
    }
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
            
            [delegate.ws publish:addedTrackMessage toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,username,self.currentPlayer.title] excludeMe:YES];
        }];
    }
    
}
@end
