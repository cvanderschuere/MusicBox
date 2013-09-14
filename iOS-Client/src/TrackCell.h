//
//  TrackCell.h
//  MusicBox
//
//  Created by Chris Vanderschuere on 4/17/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>

@interface TrackCell : UICollectionViewCell

@property (nonatomic,weak) IBOutlet UILabel* trackTitle;
@property (nonatomic,weak) IBOutlet UIImageView *albumArt;

@end
