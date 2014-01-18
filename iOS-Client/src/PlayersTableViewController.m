//
//  PlayersTableViewController.m
//  MusicBox
//
//  Created by Chris Vanderschuere on 3/30/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "PlayersTableViewController.h"

@interface PlayersTableViewController ()
@property (nonatomic,strong) NSArray* players;
@end

@implementation PlayersTableViewController

- (id)initWithStyle:(UITableViewStyle)style
{
    self = [super initWithStyle:style];
    if (self) {
        // Custom initialization
    }
    return self;
}

- (void)viewDidLoad
{
    [super viewDidLoad];

    // Uncomment the following line to preserve selection between presentations.
    // self.clearsSelectionOnViewWillAppear = NO;
 
    // Uncomment the following line to display an Edit button in the navigation bar for this view controller.
    // self.navigationItem.rightBarButtonItem = self.editButtonItem;
    
    //Create Refresh Control
    UIRefreshControl *refreshControl = [[UIRefreshControl alloc] init];
    refreshControl.tintColor = [UIColor darkGrayColor];
    [refreshControl addTarget:self action:@selector(refreshTableView:) forControlEvents:UIControlEventValueChanged];
    self.refreshControl = refreshControl;
    
    //Init players array
    self.players = [NSArray array];
    
    [self refreshTableView:self.refreshControl];
}


- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - Data methods
- (void) refreshTableView:(UIRefreshControl*)sender{
    [sender beginRefreshing];
    
    AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [delegate.websocketRequestQueue addOperationWithBlock:^{
        [delegate.ws call:[NSString stringWithFormat:@"%@players",baseURL] withDelegate:self args:nil];
    }];
}

#pragma mark Websocket Delegate
- (void) onResult:(id)result forCalledUri:(NSString *)callUri{
    if ([callUri isEqualToString:[NSString stringWithFormat:@"%@players",baseURL]]) {
        if ([result isKindOfClass:[NSArray class]]) {
            //NSLog(@"Result: %@", result);
            NSArray* boxIDs = (NSArray*) result;
            
            AppDelegate* delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
            [delegate.websocketRequestQueue addOperationWithBlock:^{
                [delegate.ws call:[NSString stringWithFormat:@"%@boxDetails",baseURL] withDelegate:self args:boxIDs, nil];
            }];
        }
    }
    else if([callUri isEqualToString:[NSString stringWithFormat:@"%@boxDetails",baseURL]]){
        //NSLog(@"Result: %@", result);
        
        NSMutableArray* playerObjs = [NSMutableArray array];
        for (NSString *key in result) {
            [playerObjs addObject:[MusicBox musicBoxWithDictionary:result[key]]];
        }
        
        self.players = playerObjs;
        
        //Use this information to populate table
        [self.tableView reloadData];
        [self.refreshControl endRefreshing];

    }
}
- (void) onError:(NSString *)errorUri description:(NSString*)errorDesc forCalledUri:(NSString *)callUri{
    if ([callUri isEqualToString:[NSString stringWithFormat:@"%@players",baseURL]]) {
        NSLog(@"Error: %@",errorDesc);
        [self.refreshControl endRefreshing];
    }
}

#pragma mark - Table view data source

- (NSInteger)numberOfSectionsInTableView:(UITableView *)tableView
{
    // Return the number of sections.
    return 1;
}

- (NSInteger)tableView:(UITableView *)tableView numberOfRowsInSection:(NSInteger)section
{
    // Return the number of rows in the section.
    return self.players.count;
}

- (UITableViewCell *)tableView:(UITableView *)tableView cellForRowAtIndexPath:(NSIndexPath *)indexPath
{
    static NSString *CellIdentifier = @"playerCell";
    UITableViewCell *cell = [tableView dequeueReusableCellWithIdentifier:CellIdentifier forIndexPath:indexPath];
    
    // Configure the cell...
    cell.textLabel.text = [self.players[indexPath.row] DeviceName];
    
    return cell;
}

#pragma mark - Table view delegate

- (void)tableView:(UITableView *)tableView didSelectRowAtIndexPath:(NSIndexPath *)indexPath
{
    [tableView deselectRowAtIndexPath:indexPath animated:YES];
    self.selectedPlayer = self.players[indexPath.row];
    [self performSegueWithIdentifier:@"unwindSeque" sender:nil];
}


- (IBAction)cancelSelection:(id)sender {
    self.selectedPlayer = nil;
    [self performSegueWithIdentifier:@"unwindSeque" sender:nil];
}
@end
