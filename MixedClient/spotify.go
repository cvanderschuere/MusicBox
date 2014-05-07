package main

import(
    "github.com/cvanderschuere/turnpike"
	"code.google.com/p/portaudio-go/portaudio"
	"github.com/cvanderschuere/go-libspotify/spotify"
	"io/ioutil"
	"fmt"

)

type spotifyClient struct{
	session *spotify.Session
	EOT chan bool
	consumer *portAudio
}

func SetupSpotify(client *turnpike.Client) *spotifyClient{
    sClient := new(spotifyClient)

    // Request pandora info
    resp := client.Call(baseURL+"userInfo",musicBoxID)

	portaudio.Initialize()

	appKey, err := ioutil.ReadFile("spotify_appkey.key")
	if err != nil {
		log.Fatal(err.Error())
	}

	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   appKey,
		ApplicationName:  "Example",
		CacheLocation:    "tmp",
		SettingsLocation: "tmp",

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	})

	if err != nil {
		log.Fatal(err.Error())
	}

    message := <-resp
    user := message.Result.(map[string]interface{})

	credentials := spotify.Credentials{
		Username: user["Username"].(string),
		Password: user["SpotifyPassword"].(string),
	}
	if err = session.Login(credentials, false); err != nil {
		log.Fatal(err.Error())
	}

	// Wait for login
	err = <-session.LoginUpdates()
	if err != nil {
		log.Fatal(err.Error())
	}

	sClient.session = session

    return sClient
}

func (c *spotifyClient)Play(){
	c.consumer.currStream.Start()
    c.session.Player().Play()
}

func (c *spotifyClient)Pause(){
	c.consumer.currStream.Stop()
    c.session.Player().Pause()
}

func (c *spotifyClient)Stop(){
	close(c.consumer.buffer)
    c.session.Player().Unload()
}

func (c *spotifyClient)SetVolume(vol uint8){
	// Do nothing for now
}

func (c *spotifyClient)NextTrack(t TrackItem) (<-chan bool){
	// Check if we need to make new consumer
	if c.consumer == nil || c.consumer.currStream == nil{
		c.consumer = newPortAudio()
		go c.consumer.player()
		c.session.SetAudioConsumer(c.consumer)

		c.EOT = c.consumer.eotChan
	}

    //Send startedTrack message
    msg := map[string]interface{} {
        "command":"startedTrack",
        "data": map[string]interface{}{
            "deviceID":musicBoxID,
            "track":t,
        },
    }
    client.PublishExcludeMe(baseURL+boxUsername+"/"+musicBoxID,msg) //Let others know track has started playing

	// Parse the track
	link, err := c.session.ParseLink(t.ProviderID)
	if err != nil {
		log.Fatal(err.Error())
	}
	track, errT := link.Track()
	if errT != nil {
		log.Fatal(errT.Error())
	}

	// Load the track and play it
	track.Wait()
	player := c.session.Player()
	if err := player.Load(track); err != nil {
		fmt.Println("%#v", err)
		log.Fatal(err.Error())
	}

	player.Play()

    return c.EOT
}

func (c *spotifyClient)ShutDown(){
	// Close portaudio
    portaudio.Terminate()

	//Close EOT chan
    if(c.EOT != nil){
	   close(c.EOT)
    }

    //Logout of services
    c.session.Close()
}

//
// PortAudio Setup
//

type audio struct {
	format spotify.AudioFormat
	frames []byte
}

type portAudio struct {
	buffer chan *audio
	eotChan chan bool
	currStream *portaudio.Stream
}

func newPortAudio() *portAudio {
	return &portAudio{
		buffer: make(chan *audio, 8),
		eotChan: make(chan bool, 2),
	}
}

func (pa *portAudio) WriteAudio(format spotify.AudioFormat, frames []byte) int {
	audio := &audio{format, frames}

	if len(frames) == 0 {
		return 0
	}

	select {
	case pa.buffer <- audio:
		return len(frames)
	default:
		return 0
	}
}

func (pa *portAudio) player() {
	out := make([]int16, 2048*2)

	var err error

	pa.currStream, err = portaudio.OpenDefaultStream(
		0,
		2,     // audio.format.Channels,
		44100, // float64(audio.format.SampleRate),
		len(out),
		&out,
	)
	if err != nil {
		panic(err)
	}

	defer func(){
		pa.currStream = nil
	}()

	defer pa.currStream.Close()

	pa.currStream.Start()
	defer pa.currStream.Stop()

	// Decode the incoming data which is expected to be 2 channels and
	// delivered as int16 in []byte, hence we need to convert it.
	for audio := range pa.buffer {
		if len(audio.frames) != 2048*2*2 {
			panic("unexpected")
		}

		j := 0
		for i := 0; i < len(audio.frames); i += 2 {
			out[j] = int16(audio.frames[i]) | int16(audio.frames[i+1])<<8
			j++
		}

		pa.currStream.Write()
	}
	fmt.Println("Exited Player loop")

	pa.eotChan<-true
}
