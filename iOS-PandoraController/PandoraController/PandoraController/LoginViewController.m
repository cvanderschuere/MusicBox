//
//  LoginViewController.m
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import "LoginViewController.h"
#import "AppDelegate.h"
#import "User.h"
#import "MainViewController.h"

@interface LoginViewController ()

@end

@implementation LoginViewController

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
    
}

- (void) viewWillAppear:(BOOL)animated{    
    
    //Connect to websocket
    AppDelegate *delegate = (AppDelegate*)[UIApplication sharedApplication].delegate;
    [delegate connectToWebSocket];
    
    //
    // Update screen
    //
    
    [self updateBuildInformation];
    
    //Add login information for now
    self.usernameTextField.text = @"christopher.vanderschuere@gmail.com";
    self.passwordTextField.text = @"testPassword";
}

- (void) updateBuildInformation{
    //Load Build Information
	NSDictionary *infoDictionary = [[NSBundle mainBundle] infoDictionary];
	NSString *name = [infoDictionary objectForKey:@"CFBundleDisplayName"];
	NSString *version = [infoDictionary objectForKey:@"CFBundleShortVersionString"];
	NSString *build = [infoDictionary objectForKey:@"CFBundleVersion"];
	self.buildLabel.text = [NSString stringWithFormat:@"%@ v%@ (build %@)",name,version,build];
}

- (void)didReceiveMemoryWarning
{
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

#pragma mark - IBActions
- (IBAction)loginTapped:(UIButton *)sender {
    //Use login information to login websocket
    NSString *username = self.usernameTextField.text;
    NSString *password = self.passwordTextField.text;
    
    if ([username length] == 0 || [password length] == 0) {
        return; //Error
    }
    
    AppDelegate * delegate = (AppDelegate*) [UIApplication sharedApplication].delegate;
    [delegate.ws authenticateWebsocketWithUsername:username Password:password Callback:^(User* user, NSError* error) {
        [self.loadingIndicator stopAnimating];
        if (error == nil) {
            NSLog(@"Successfully logged in");
            
            //Logged in...move to next screen
            [self performSegueWithIdentifier:@"loginSegue" sender:user];
            
        }else{
            //Do something with the error
            NSLog(@"Error:%@",error);
        }
        
    }];
}

#pragma mark - Segue
- (void) prepareForSegue:(UIStoryboardSegue *)segue sender:(id)sender{
    if ([segue.identifier isEqualToString:@"loginSegue"]) {
        User* user = (User*) sender;
        
        //Prepare destination with user
        MainViewController* mainVC = (MainViewController*) segue.destinationViewController;
        [mainVC setCurrentUser:user];
    }
    
}
@end
