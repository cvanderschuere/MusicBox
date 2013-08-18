//
//  TrackSearchViewController.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "TrackSearchViewController.h"
#import "SpotifyTrack.h"

@interface TrackSearchViewController()

@property (nonatomic,strong) NSMutableArray* resultsArray;
@property (nonatomic, strong) RKResponseDescriptor *trackResponseDescriptor;

@end

@implementation TrackSearchViewController

- (void)viewDidLoad
{
    [super viewDidLoad];
    
    //FIXME: allow scope changes
    self.searchBar.selectedScopeButtonIndex = 2;
    self.searchBar.showsScopeBar = NO;
    
    //Create object mappings
    
    //Album
    RKObjectMapping *albumMapping = [RKObjectMapping mappingForClass:[SpotifyAlbum class]];
    [albumMapping addAttributeMappingsFromDictionary:@{@"name": @"name",@"href":@"spotifyID"}];
    
    //Artist
    RKObjectMapping *artistMapping = [RKObjectMapping mappingForClass:[SpotifyArtist class]];
    [artistMapping addAttributeMappingsFromDictionary:@{@"name": @"name",@"href":@"spotifyID"}];

    
    //Track
    RKObjectMapping *trackMapping = [RKObjectMapping mappingForClass:[SpotifyTrack class]];
    [trackMapping addAttributeMappingsFromDictionary:@{@"name": @"name",@"href":@"spotifyID"}];
    [trackMapping addRelationshipMappingWithSourceKeyPath:@"album" mapping:albumMapping];
    [trackMapping addRelationshipMappingWithSourceKeyPath:@"artists" mapping:artistMapping];
    

    //Response Descriptor
    NSIndexSet *statusCodes = RKStatusCodeIndexSetForClass(RKStatusCodeClassSuccessful); // Anything in 2xx
    self.trackResponseDescriptor = [RKResponseDescriptor responseDescriptorWithMapping:trackMapping method:RKRequestMethodAny pathPattern:nil keyPath:@"tracks" statusCodes:statusCodes];
    
}
- (void) dealloc{
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}
#pragma mark - UITableView Delegate
- (void) tableView:(UITableView *)tableView didSelectRowAtIndexPath:(NSIndexPath *)indexPath{
    [tableView deselectRowAtIndexPath:indexPath animated:YES];
    
    SpotifyTrack* track = self.resultsArray[indexPath.row];
    
    
    self.selectedTrack = [MusicBoxTrack trackWithService:@"Spotify" Url:track.spotifyID Name:track.name Album:track.album.name Artist:[track.artists[0] name]];
    
    [self.searchBar resignFirstResponder];
    [self performSegueWithIdentifier:@"unwindSegue" sender:self];
}
#pragma mark - UITableView Datasource
-(NSInteger) tableView:(UITableView *)tableView numberOfRowsInSection:(NSInteger)section{
    return self.resultsArray.count;
}
-(UITableViewCell*) tableView:(UITableView *)tableView cellForRowAtIndexPath:(NSIndexPath *)indexPath{
    UITableViewCell* cell = [tableView dequeueReusableCellWithIdentifier:@"SearchCell" forIndexPath:indexPath];
    
    
    NSString *title = nil;
    NSString *subtitle = nil;
    switch (self.searchBar.selectedScopeButtonIndex) {
        case 0:
            title = @"Artist";
            break;
        case 1: //Album
            title = @"Album";
            break;
        case 2: //Track
        {
            SpotifyTrack *track = self.resultsArray[indexPath.row];
            
            title = track.name;
            subtitle = track.album.name;
            break;
        }
        default:
            break;
    }
    
    cell.textLabel.text = title;
    cell.detailTextLabel.text = subtitle;
    return cell;
}

#pragma mark - UISearchBar Delegate

-(void)searchBar:(UISearchBar *)searchBar selectedScopeButtonIndexDidChange:(NSInteger)selectedScope{
    //[self.results reloadSections:[NSIndexSet indexSetWithIndex:0] withRowAnimation:UITableViewRowAnimationAutomatic];
}
- (void) searchBarCancelButtonClicked:(UISearchBar *)searchBar{
    [searchBar resignFirstResponder];
}
- (void) searchBarSearchButtonClicked:(UISearchBar *)searchBar{
    [searchBar resignFirstResponder];
}
- (BOOL) searchBarShouldBeginEditing:(UISearchBar *)searchBar{
    [searchBar setShowsCancelButton:YES animated:YES];
    return YES;

}
- (BOOL) searchBarShouldEndEditing:(UISearchBar *)searchBar{
    [searchBar setShowsCancelButton:NO animated:YES];
    return YES;
}
- (void) searchBar:(UISearchBar *)searchBar textDidChange:(NSString *)searchText{
    if (searchText.length == 0)
		return;
    
    //Escape search string
    NSString *escapedString = (NSString *)CFBridgingRelease(CFURLCreateStringByAddingPercentEscapes(
                                                                                                    NULL,
                                                                        (__bridge CFStringRef) searchText,
                                                                                                    NULL,
                                                                            CFSTR("!*'();:@&=+$,/?%#[]"),
                                                                                kCFStringEncodingUTF8));
	//TODO: Perform search with this information
    NSURLRequest *request = [NSURLRequest requestWithURL:[NSURL URLWithString:[NSString stringWithFormat:@"http://ws.spotify.com/search/1/track.json?q=%@",escapedString]]];
    RKObjectRequestOperation *operation = [[RKObjectRequestOperation alloc] initWithRequest:request responseDescriptors:@[self.trackResponseDescriptor]];
    [operation setCompletionBlockWithSuccess:^(RKObjectRequestOperation *operation, RKMappingResult *result) {
        self.resultsArray = result.array.mutableCopy;
        [self.results reloadData];
    } failure:^(RKObjectRequestOperation *operation, NSError *error) {
        NSLog(@"Failed with error: %@", [error localizedDescription]);
    }];
    [operation start];
    
}
@end
