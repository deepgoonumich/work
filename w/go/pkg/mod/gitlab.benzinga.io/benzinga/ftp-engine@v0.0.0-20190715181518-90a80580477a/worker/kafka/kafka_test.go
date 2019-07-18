package kafka

import (
	"context"
	"os"
	"testing"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	"github.com/icrowley/fake"
	jsoniter "github.com/json-iterator/go"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"

	"gitlab.benzinga.io/benzinga/bzkaf"
	"gitlab.benzinga.io/benzinga/content-models/models"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/ftp-engine/instr"
	"gitlab.benzinga.io/benzinga/ftp-engine/process/ravenpack"
	"gitlab.benzinga.io/benzinga/ftp-engine/rstore"
	"gitlab.benzinga.io/benzinga/ftp-engine/sender/ftp"
)

func TestKafka(t *testing.T) {
	// Test Setup
	cfg, err := config.LoadConfig("testing")
	require.NoError(t, err)

	logger, err := cfg.LoadLogger()
	require.NoError(t, err)

	// Start Test FTP Server
	factory := &filedriver.FileDriverFactory{
		RootPath: os.TempDir(),
		Perm:     server.NewSimplePerm("user", "group"),
	}

	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     12345,
		Hostname: "127.0.0.1",
		Auth:     &server.SimpleAuth{Name: cfg.FTP.Username, Password: cfg.FTP.Password},
	}

	ftpServer := server.NewServer(opts)
	go func() {
		require.NoError(t, ftpServer.ListenAndServe())
		assert.NoError(t, ftpServer.Shutdown())
	}()

	cfg.FTP.Host = "localhost:12345"

	s, err := ftp.NewFTPSender(cfg, logger)
	require.NoError(t, err)

	inst, err := instr.NewCollector(cfg.AppName)
	require.NoError(t, err)

	rClient, err := rstore.NewClient(logger, cfg.RedisURL)
	require.NoError(t, err)

	processor := ravenpack.NewRavenpackProcessor(cfg, rClient, logger)

	w, err := NewKafkaWorker(cfg, logger, inst, s, processor)
	require.NoError(t, err, "Load Kafka Worker Error")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	// Start Pushing Test Events
	loadTestKafkaContent(ctx, t, cfg, 5)

	w.Work(ctx)

	cancel()
	assert.NoError(t, ftpServer.Shutdown())

}

func loadTestKafkaContent(ctx context.Context, t *testing.T, cfg *config.Config, testEvents int) {

	kafkaConfig := kafka.WriterConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	}

	require.NoError(t, kafkaConfig.Validate())

	w := kafka.NewWriter(kafkaConfig)

	for i := 0; i < testEvents; i++ {

		content, err := jsoniter.Marshal(newTestEvent())
		require.NoError(t, err, "Marshal Content Error")
		// New Envelope
		envelope := bzkaf.NewEnvelope(bzkaf.ContentModelsEventMsgType, content)
		envelopeJSON, err := envelope.Marshal()
		require.NoError(t, err, "Marshal Envelope Error")

		// Create Message
		msg := kafka.Message{
			Value: envelopeJSON,
		}

		require.NoError(t, w.WriteMessages(ctx, msg), "Write Message Error")

		t.Log("Message Write Success")

	}

	assert.NoError(t, w.Close(), "Kafka Writer Close Error")

}

func newTestEvent() *models.Event {

	id := int64(randomdata.Number(9999999999999))
	node := int64(randomdata.Number(9999999999999))
	event := models.Event{
		ID:     id,
		NodeID: node,
		Time: models.Time{
			Time: time.Now(),
		},
		Content: models.Content{
			ID:        bson.ObjectIdHex("50de2127c3be26ca32000000"),
			Title:     fake.Sentence(),
			Body:      testBody,
			Author:    fake.FullName(),
			NodeID:    int(node),
			VersionID: int(id),
			EventID:   bson.ObjectIdHex("50de2127c3be26ca32000000"),
			Published: true,
			Type:      "story",
			CreatedAt: models.Time{
				Time: time.Now(),
			},
			UpdatedAt: models.Time{
				Time: time.Now(),
			},
			Tickers: []models.Category{
				{ID: 14502, Vocab: 2, Name: "F", Description: "Ford Motor Credit Company", Primary: true},
				{ID: 42010, Vocab: 2, Name: "GLOG", Description: "", Primary: false},
			},
			Channels: []models.Category{
				{ID: 57, Vocab: 1, Name: "News"},
			},
			PartnerURL: "test",
			IsBzPost:   true,
			Meta: models.Meta{
				SectorV2: &models.SectorMeta{
					SIC: []models.SICSector{
						{
							IndustryCode: 1234,
						},
					},
				},
				Ext: map[string]interface{}{
					"Test": map[string]interface{}{
						"TestField": 1,
					},
				},
			},
			Assets: []models.Asset{
				{
					Type:    models.ImageAsset,
					Primary: true,
					MIME:    "image/jpeg",
					Attributes: &models.AssetAttributes{
						Filename: "qualcomm-hq-5-web_1.jpg",
						Filepath: "files/images/story/2012/qualcomm-hq-5-web_1.jpg",
					},
				},
			},
		},
		Event: models.Created,
	}

	return &event
}

const testBody = `
<p>Liquefied natural gas (LNG) shipping&#39;s first and only cooperative spot pool is disbanding, a move that should bring increased competition to the marketplace &ndash; at least, until a replacement or reincarnation emerges.</p>

<p>The Cool Pool was launched in September 2015 with 14 tri-fuel diesel engine (TFDE) LNG carriers contributed by three owners: three ships from Dynagas Ltd, three from <strong>GasLog Ltd</strong> (NYSE:<a class="ticker" href="/stock/GLOG#NYSE">GLOG</a>) and eight from <strong>Golar LNG Ltd</strong> (NASDAQ:<a class="ticker" href="/stock/GLNG#NASDAQ">GLNG</a>).</p>

<p>Each participant retained technical management of its own ships, but the Cool Pool manager coordinated all employment of one year or less in duration; if an owner secured longer-term employment for one of its participating vessels, it would be removed from the pool.</p>

<p>The idea of the Cool Pool was to improve the utilization of the participating vessels (the time spent laden versus in ballast) and to allow for the use of Contracts of Affreightment (COAs). In a COA, a cargo shipper does not employ a specific vessel, it contracts for the carriage of a certain volume over a specified time period. Servicing a COA requires having enough ships and enough flexibility &ndash; something an individual ship owner would not possess in the LNG spot business, but a co-operative pool would.</p>

<p>The pool also allowed Dynagas, GasLog and Golar to do business with a much wider array of clients, allowing those clients to become familiar with the owners and potentially do future business with them on a long-term basis. In other words, the Cool Pool was good marketing.</p>

<p>There was also a potential pricing advantage. The number of ships in the global LNG spot market fluctuated, but at times, the Cool Pool fleet accounted for more than a third of all spot LNG vessels on the water, leading some financial analysts to speculate that the Cool Pool had at least some ability to drive up LNG freight rates.</p>

<p>Executives of participating ship owners confirmed on conference calls over recent years that the concept was a success. What ultimately sank the Cool Pool &ndash; or at least, its first incarnation &ndash; was not that the idea didn&#39;t work, but that the business strategies of its founding members diverged.</p>

<p>In mid-2018, Dynagas secured long-term employment for all three of its participating ships and pulled out of the Cool Pool.</p>

<p>The next development came on May 21, when Golar LNG Ltd reported its quarterly earnings. The company currently has three business focuses: floating liquefaction, the power sector, and LNG shipping. It confirmed in its quarterly release that its <a href="https://www.freightwaves.com/news/more-ship-owners-head-to-wall-street-via-direct-listings-not-ipos" rel="noreferrer noopener" target="_blank">LNG shipping business would be spun off into a separate &lsquo;pure play&#39; listed entity</a>.</p>

<p>On June 6, GasLog Ltd and <strong>GasLog Partners</strong> (NYSE:<a class="ticker" href="/stock/GLOP#NYSE">GLOP</a>) announced that they would remove all ships from the Cool Pool and retake commercial control &quot;over the coming months&quot; (GasLog Ltd has six ships in the Cool Pool; GasLog Partners has one).</p>

<p>The GasLog companies said that their decision was in response to Golar LNG&#39;s move to spin off its shipping business into a separately listed company, as well as their belief in improving fundamentals.</p>

<p>According to Paul Wogan, chief executive officer of GasLog Ltd, &quot;With Golar&#39;s declared intention to spin off its LNG vessels and a tightening of the LNG carrier market now underway, we believe it is the right time to assume control of our vessel marketing as we seek to place more vessels on longer-term charters to optimize the earnings of our fleet through the cycle. This move is underpinned by increasing levels of customer enquiry in multi-month and multi-year charters.&quot;</p>

<p>Within hours of GasLog&#39;s statement, Golar LNG Ltd responded. &quot;Golar, or subject to interim market conditions, the spin-off entity, will assume ownership of the Cool Pool following GasLog&#39;s departure. There will be a ramp-down period to allow for the conclusion of existing GasLog vessel charter contracts.&quot;</p>

<p>A pool, by definition, is a cooperative arrangement between more than one ship owner; when GasLog leaves, the Cool Pool will technically not function as a cooperative entity. Yet Golar&#39;s statement refers to &quot;changes&quot; for the pool, not its demise, and it suggests a path forward for its revival with new members.</p>

<p>It stated that the ships in its separately listed LNG shipping company are &quot;expected to continue to trade within the Cool Pool after a formal launch of the spin-off. Transfer of the Cool Pool to this new shipping entity is expected to create the leading independent provider of available on-the-water TFDE LNG carriers.&quot;</p>

<p>It also stated that it is &quot;in talks with other owners of similar tonnage to join the new shipping entity,&quot; which could be interpreted as meaning that it is in talks with others to share in ownership in the new listed vehicle, or alternatively, to join a reborn Cool Pool &ndash; or both.</p>

<p>But the Cool Pool&#39;s resurgence is far from guaranteed, nor is it certain that all of the GasLog vessels will shift into the long-term charter market. Consequently, the latest developments could effectively create more competitors bidding for spot LNG cargoes, which would theoretically be a headwind for rates.</p>

<p>Image sourced from Pixabay</p>`
