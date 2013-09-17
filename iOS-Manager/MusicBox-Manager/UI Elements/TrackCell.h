//
//  TrackCell.h
//  MusicBox-Manager
//
//  Created by Chris Vanderschuere on 9/16/13.
//  Copyright (c) 2013 CDVConcepts. All rights reserved.
//

#import <UIKit/UIKit.h>

@interface TrackCell : UICollectionViewCell

@property (nonatomic,strong) IBOutlet UIImageView* artworkImageView;
@property (nonatomic,strong) IBOutlet UILabel* artistLabel;
@property (nonatomic,strong) IBOutlet UILabel* trackLabel;

@end
