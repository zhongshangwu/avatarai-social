package repositories

type ActivityRepository struct {
	metaStore *MetaStore
}

func NewActivityRepository(metaStore *MetaStore) *ActivityRepository {
	return &ActivityRepository{metaStore: metaStore}
}

func (r *ActivityRepository) CreateTag(tag *Tag) error {
	return r.metaStore.DB.Create(tag).Error
}

func (r *ActivityRepository) GetTagByTag(tag string) (*Tag, error) {
	var result Tag
	if err := r.metaStore.DB.Where("tag = ? AND deleted = ?", tag, false).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ActivityRepository) DeleteTag(tag string) error {
	return r.metaStore.DB.Model(&Tag{}).Where("tag = ?", tag).Update("deleted", true).Error
}

func (r *ActivityRepository) CreateActivityTag(activityTag *ActivityTag) error {
	return r.metaStore.DB.Create(activityTag).Error
}

func (r *ActivityRepository) GetActivityTagsBySubjectURI(subjectURI string) ([]*ActivityTag, error) {
	var tags []*ActivityTag
	if err := r.metaStore.DB.Where("subject_uri = ? AND deleted = ?", subjectURI, false).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *ActivityRepository) GetActivityTagsBySubjectURIs(subjectURIs []string) (map[string][]*ActivityTag, error) {
	ret := make(map[string][]*ActivityTag)
	var tags []*ActivityTag
	if err := r.metaStore.DB.Where("subject_uri IN ? AND deleted = ?", subjectURIs, false).Find(&tags).Error; err != nil {
		return nil, err
	}
	for _, tag := range tags {
		ret[tag.SubjectURI] = append(ret[tag.SubjectURI], tag)
	}
	return ret, nil
}

func (r *ActivityRepository) DeleteActivityTag(tag string, subjectURI string) error {
	return r.metaStore.DB.Model(&ActivityTag{}).Where("tag = ? AND subject_uri = ?", tag, subjectURI).Update("deleted", true).Error
}

func (r *ActivityRepository) GetTagActivityCounts(tags []string) (map[string]int, error) {
	counts := make(map[string]int)
	query := r.metaStore.DB.Model(&ActivityTag{}).Where("tag IN ? AND deleted = ?", tags, false).
		Select("tag, COUNT(*) as count").
		Group("tag")
	if err := query.Scan(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *ActivityRepository) CreateTopic(topic *Topic) error {
	return r.metaStore.DB.Create(topic).Error
}

func (r *ActivityRepository) GetTopicByTopic(topic string) (*Topic, error) {
	var result Topic
	if err := r.metaStore.DB.Where("topic = ? AND deleted = ?", topic, false).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ActivityRepository) DeleteTopic(topic string) error {
	return r.metaStore.DB.Model(&Topic{}).Where("topic = ?", topic).Update("deleted", true).Error
}

func (r *ActivityRepository) CreateActivityTopic(activityTopic *ActivityTopic) error {
	return r.metaStore.DB.Create(activityTopic).Error
}

func (r *ActivityRepository) GetActivityTopicsBySubjectURI(subjectURI string) ([]*ActivityTopic, error) {
	var topics []*ActivityTopic
	if err := r.metaStore.DB.Where("subject_uri = ? AND deleted = ?", subjectURI, false).Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *ActivityRepository) GetActivityTopicsBySubjectURIs(subjectURIs []string) (map[string][]*ActivityTopic, error) {
	ret := make(map[string][]*ActivityTopic)
	var topics []*ActivityTopic
	if err := r.metaStore.DB.Where("subject_uri IN ? AND deleted = ?", subjectURIs, false).Find(&topics).Error; err != nil {
		return nil, err
	}
	for _, topic := range topics {
		ret[topic.SubjectURI] = append(ret[topic.SubjectURI], topic)
	}
	return ret, nil
}

func (r *ActivityRepository) DeleteActivityTopic(topic string, subjectURI string) error {
	return r.metaStore.DB.Model(&ActivityTopic{}).Where("topic = ? AND subject_uri = ?", topic, subjectURI).Update("deleted", true).Error
}

func (r *ActivityRepository) GetTopicActivityCounts(topics []string) (map[string]int, error) {
	counts := make(map[string]int)
	query := r.metaStore.DB.Model(&ActivityTopic{}).Where("topic IN ? AND deleted = ?", topics, false).
		Select("topic, COUNT(*) as count").
		Group("topic")
	if err := query.Scan(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *ActivityRepository) ListTags(page int, pageSize int) ([]*Tag, error) {
	var tags []*Tag
	query := r.metaStore.DB.Where("deleted = ?", false).Order("created_at DESC")

	if page > 0 {
		query = query.Offset((page - 1) * pageSize)
	}

	if pageSize > 0 {
		query = query.Limit(pageSize)
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *ActivityRepository) ListTopics(page int, pageSize int) ([]*Topic, error) {
	var topics []*Topic
	query := r.metaStore.DB.Where("deleted = ?", false).Order("created_at DESC")

	if page > 0 {
		query = query.Offset((page - 1) * pageSize)
	}

	if pageSize > 0 {
		query = query.Limit(pageSize)
	}

	if err := query.Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}
