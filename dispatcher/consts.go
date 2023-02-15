package dispatcher

import "time"

const DefaultAnnouncements = "announcements"

const DefaultConsumersTTL = 30 * time.Second
const DefaultConsumersCheckTTLPeriod = DefaultConsumersTTL / 2

const DefaultDispatcherPeriod = 1 * time.Minute

const DefaultAnnouncementPeriod = time.Second
