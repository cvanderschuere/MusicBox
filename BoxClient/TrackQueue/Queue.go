package TrackQueue

import(
	"MusicBox/BoxClient/Track"
)

type Queue struct{
	Tracks []Track.Track
	History []Track.Track
}

func NewQueue(tracks []Track.Track){
	q := new(Queue);
	q.Tracks = make([]Track.Track, 0)
	q.History = make([]Track.Track, 100)
	
	if(tracks != nil){
		for track := range(tracks)
		append(q.Tracks, tracks);
	}
}

func (q *Queue) AddTrack(track Track.Track){
	append(q.Tracks, track);
}

func (q *Queue) AddTracks(tracks []Track.Track){
	for track := range(tracks){
		q.AddTrack(track);
	}
}

func (q *Queue) NextTrack()(Track.Track){
	q.History = q.History[99:];
	append(q.History, q.Tracks[0])
	
	q.Tracks = q.Tracks[1:];
	
	return q.Tracks[0];
}

func (q *Queue) InsertSong(track Track.Track, newPosition int){
	
}