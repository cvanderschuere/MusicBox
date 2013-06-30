//
//  TrackSearchViewController.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/28/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "TrackSearchViewController.h"
#import "LiveSearch.h"

@interface TrackSearchViewController ()
@property (nonatomic, strong) LiveSearch* liveSearch;
@end

@implementation TrackSearchViewController

- (void)viewDidLoad
{
    [super viewDidLoad];
    
    //Add observer for live search
    [self addObserver:self forKeyPath:@"liveSearch.latestSearch.loaded" options:0 context:NULL];
}
- (void) dealloc{
    [self removeObserver:self forKeyPath:@"liveSearch.latestSearch.loaded"];
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}
- (void) observeValueForKeyPath:(NSString *)keyPath ofObject:(id)object change:(NSDictionary *)change context:(void *)context{
    if ([keyPath hasPrefix:@"liveSearch"]) {
        //Live search updated
        
        //Pick selected index based on top hit
        if ([self.liveSearch.topHit isKindOfClass:[SPArtist class]]) {
            [self.searchBar setSelectedScopeButtonIndex:0];
        }
        else if ([self.liveSearch.topHit isKindOfClass:[SPAlbum class]]){
            [self.searchBar setSelectedScopeButtonIndex:1];
        }
        else if([self.liveSearch.topHit isKindOfClass:[SPTrack class]]){
            [self.searchBar setSelectedScopeButtonIndex:2];
        }
        
        //Update tableview accordingly
        [self.results reloadData];
    }
    else
        [super observeValueForKeyPath:keyPath ofObject:object change:change context:context];
}
#pragma mark - UITableView Delegate
- (void) tableView:(UITableView *)tableView didSelectRowAtIndexPath:(NSIndexPath *)indexPath{
    [tableView deselectRowAtIndexPath:indexPath animated:YES];
    
    NSURL* url = nil;
    switch (self.searchBar.selectedScopeButtonIndex) {
        case 0: //Artist
            url = [[self.liveSearch.topArtists objectAtIndex:indexPath.row] spotifyURL];
            break;
        case 1: //Album
            url = [[self.liveSearch.topAlbums objectAtIndex:indexPath.row] spotifyURL];
            break;
        case 2: //Track
            url = [[self.liveSearch.topTracks objectAtIndex:indexPath.row] spotifyURL];
            break;
        default:
            break;
    }
    
    self.selectedTrackURL = url;

    
    [self.searchBar resignFirstResponder];
    [self performSegueWithIdentifier:@"unwindSegue" sender:self];
}
#pragma mark - UITableView Datasource
-(NSInteger) tableView:(UITableView *)tableView numberOfRowsInSection:(NSInteger)section{
    switch (self.searchBar.selectedScopeButtonIndex) {
        case 0: //Artist
            return self.liveSearch.topArtists.count;
            break;
        case 1: //Album
            return self.liveSearch.topAlbums.count;
            break;
        case 2: //Track
            return self.liveSearch.topTracks.count;
            break;
        default:
            return 0;
            break;
    }
}
-(UITableViewCell*) tableView:(UITableView *)tableView cellForRowAtIndexPath:(NSIndexPath *)indexPath{
    UITableViewCell* cell = [tableView dequeueReusableCellWithIdentifier:@"SearchCell" forIndexPath:indexPath];
    NSString *title = nil;
    switch (self.searchBar.selectedScopeButtonIndex) {
        case 0: //Artist
            title = [[self.liveSearch.topArtists objectAtIndex:indexPath.row] name];
            break;
        case 1: //Album
            title = [[self.liveSearch.topAlbums objectAtIndex:indexPath.row] name];
            break;
        case 2: //Track
            title = [[self.liveSearch.topTracks objectAtIndex:indexPath.row] name];
            break;
        default:
            break;
    }
    
    cell.textLabel.text = title;
    return cell;
}

#pragma mark - UISearchBar Delegate

-(void)searchBar:(UISearchBar *)searchBar selectedScopeButtonIndexDidChange:(NSInteger)selectedScope{
    [self.results reloadData];
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
    if ([searchText isEqualToString:self.liveSearch.latestSearch.searchQuery] || searchText.length == 0)
		return;
	
	SPSearch *newSearch = [SPSearch liveSearchWithSearchQuery:searchText
													inSession:[SPSession sharedSession]];
	
	if (self.liveSearch == nil) {
		self.liveSearch = [[LiveSearch alloc] initWithInitialSearch:newSearch];
	} else {
		self.liveSearch.latestSearch = newSearch;
	}

}
@end
