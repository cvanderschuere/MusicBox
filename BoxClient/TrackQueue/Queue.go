package TrackQueue

import(
	"MusicBox/BoxClient/Track"
)

type Queue struct{
	Tracks []Track.Track
	History []Track.Track
}

func NewQueue(tracks []Track.Track) (*Queue){
	q := new(Queue);
	q.Tracks = make([]Track.Track, 0)
	q.History = make([]Track.Track, 100)
	
	if(tracks != nil){
		for i := range(tracks){
			q.Tracks = append(q.Tracks, tracks[i]);
		}
	}
	
	return q;
}

func (q *Queue) AddTrack(track Track.Track){
	q.Tracks = append(q.Tracks, track);
}

func (q *Queue) AddTracks(tracks []Track.Track){
	for i := range(tracks){
		q.AddTrack(tracks[i]);
	}
}

func (q *Queue) NextTrack()(Track.Track){
	q.History = q.History[99:];
	q.History = append(q.History, q.Tracks[0])
	
	q.Tracks = q.Tracks[1:];
	
	return q.Tracks[0];
}

func (q *Queue) InsertSong(track Track.Track, newPosition int){
	
}
