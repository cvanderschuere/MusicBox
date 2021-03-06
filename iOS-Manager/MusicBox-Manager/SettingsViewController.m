//
//  SettingsViewController.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/22/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "SettingsViewController.h"
#import "AppDelegate.h"
#import "SelectThemeViewController.h"

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
    //self.navigationItem.rightBarButtonItem = self.editButtonItem;
    
    
    self.themeLabel.text = self.selectedDevice.themeObj[@"Name"];
    self.deviceNameLabel.text = self.selectedDevice.deviceName;
    
}

- (void) viewWillAppear:(BOOL)animated{
    [super viewWillAppear:animated];
    
    [self.selectedDevice addObserver:self forKeyPath:@"playState" options: NSKeyValueObservingOptionInitial|NSKeyValueObservingOptionNew context:NULL];
    //Inital state sent before end of add observer
}

- (void) viewWillDisappear:(BOOL)animated{
    [super viewWillDisappear:animated];
    
    [self.selectedDevice removeObserver:self forKeyPath:@"playState"];
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}


#pragma mark - Navigation

- (void) unwindFromThemeSelection:(UIStoryboardSegue *)segue{
    SelectThemeViewController* selectVC = (SelectThemeViewController*) segue.sourceViewController;
    
    if (selectVC.selectedTheme) {
        self.selectedDevice.themeObj = selectVC.selectedTheme;
        self.selectedDevice.theme = selectVC.selectedTheme[@"ThemeID"];
        self.themeLabel.text = selectVC.selectedTheme[@"Name"];
        
        //Publish notice of theme change
        AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
        [del.ws publish:@{
                          @"command":@"updateTheme",
                          @"data":self.selectedDevice.themeObj
                          }
                toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,self.selectedDevice.user,self.selectedDevice.identifier]];
    }
    
}

// In a story board-based application, you will often want to do a little preparation before navigation
- (void)prepareForSegue:(UIStoryboardSegue *)segue sender:(id)sender
{
    if ([segue.identifier isEqualToString:@"editTheme"]) {
        SelectThemeViewController* selectVC = (SelectThemeViewController*) segue.destinationViewController;
        selectVC.selectedTheme = self.selectedDevice.themeObj;
        
    }
    
}

- (IBAction)playerStatusChanged:(UIBarButtonItem *)sender {
    AppDelegate *del = (AppDelegate*) [UIApplication sharedApplication].delegate;
    
    if (sender.tag == 0) {
        //Play player
        [del.ws publish:@{
                          @"command":@"playTrack",
                          }
                toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,self.selectedDevice.user,self.selectedDevice.identifier]];
        
        self.selectedDevice.playState = [NSNumber numberWithInt:DEVICE_PLAYING];
        
    }else{
        //Pause player
        [del.ws publish:@{
                          @"command":@"pauseTrack",
                          }
                toTopic:[NSString stringWithFormat:@"%@%@/%@",baseURL,self.selectedDevice.user,self.selectedDevice.identifier]];
        
        
        self.selectedDevice.playState = [NSNumber numberWithInt:DEVICE_PAUSED];
    }
    
}

#pragma mark - Key value observing
- (void) observeValueForKeyPath:(NSString *)keyPath ofObject:(id)object change:(NSDictionary *)change context:(void *)context{
    
    //Handle device play state
    if ([keyPath isEqualToString:@"playState"]) {
        //Update UI to new isPlaying state
        switch (self.selectedDevice.playState.intValue) {
            case DEVICE_PAUSED:
            {
                //Create play button
                UIBarButtonItem *playItem = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPlay target:self action:@selector(playerStatusChanged:)];
                playItem.tintColor = [UIColor greenColor];
                playItem.tag = 0;
                
                self.navigationItem.rightBarButtonItem = playItem;
                
                break;
            }
            case DEVICE_PLAYING:
            {
                //Create pause button
                UIBarButtonItem *pauseItem = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemPause target:self action:@selector(playerStatusChanged:)];
                pauseItem.tintColor = [UIColor redColor];
                pauseItem.tag = 1;
                self.navigationItem.rightBarButtonItem = pauseItem;
                break;
            }
            case DEVICE_OFFLINE: //Fallthrough
            default:
            {
                //Offline
                UIBarButtonItem *pauseItem = [[UIBarButtonItem alloc] initWithBarButtonSystemItem:UIBarButtonSystemItemStop target:self action:nil];
                pauseItem.tintColor = [UIColor blackColor];
                pauseItem.enabled = NO;
                pauseItem.tag = 1;
                self.navigationItem.rightBarButtonItem = pauseItem;
                
                break;
            }
        }
    }
    
}

@end
