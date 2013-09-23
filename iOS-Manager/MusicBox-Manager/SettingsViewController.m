//
//  SettingsViewController.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/22/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "SettingsViewController.h"

@interface SettingsViewController ()

@end

@implementation SettingsViewController

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
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}


#pragma mark - Navigation


- (IBAction)unwindFromCancelledModification:(UIStoryboardSegue* ) segue{
    [self dismissViewControllerAnimated:YES completion:nil];
}
- (IBAction)unwindFromSavedModification:(UIStoryboardSegue* ) segue{
    //Modify from changes
    
    [self dismissViewControllerAnimated:YES completion:nil];
}

// In a story board-based application, you will often want to do a little preparation before navigation
- (void)prepareForSegue:(UIStoryboardSegue *)segue sender:(id)sender
{
    // Get the new view controller using [segue destinationViewController].
    // Pass the selected object to the new view controller.
    
    if ([segue.identifier isEqualToString:@"ModifySegue"]) {

    }
    
}


- (IBAction)playerStatusChanged:(UIBarButtonItem *)sender {
    if (sender.tag == 0) {
        //Pause player
        UIBarButtonItem *pauseItem = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPause target:self action:@selector(playerStatusChanged:)];
        pauseItem.tintColor = [UIColor redColor];
        pauseItem.tag = 1;
        self.navigationItem.rightBarButtonItem = pauseItem;
        
    }else{
        //Start player
        UIBarButtonItem *playItem = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPlay target:self action:@selector(playerStatusChanged:)];
        playItem.tintColor = [UIColor greenColor];
        playItem.tag = 0;

        self.navigationItem.rightBarButtonItem = playItem;
        
    }
    
}
@end
