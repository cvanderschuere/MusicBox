//
//  SettingsViewController.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/22/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>
#import "Device.h"

@interface SettingsViewController : UITableViewController

@property (nonatomic, weak) Device *selectedDevice;

//Unwinds
- (IBAction)unwindFromCancelledModification:(UIStoryboardSegue* ) segue;
- (IBAction)unwindFromSavedModification:(UIStoryboardSegue* ) segue;

//Actions
- (IBAction)playerStatusChanged:(UIBarButtonItem *)sender;

//IBOutlets
@property (weak, nonatomic) IBOutlet UILabel *deviceNameLabel;
@property (weak, nonatomic) IBOutlet UILabel *cityLabel;
@property (weak, nonatomic) IBOutlet UILabel *moodLabel;
@property (weak, nonatomic) IBOutlet UILabel *weatherLabel;
@property (weak, nonatomic) IBOutlet UILabel *keywordLabel;
@property (weak, nonatomic) IBOutlet UISwitch *limitSongsSwitch;

@end
