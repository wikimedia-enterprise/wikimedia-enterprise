DROP STREAM IF EXISTS rlt_articles_str;
CREATE STREAM IF NOT EXISTS rlt_articles_str_v2 WITH (
  KAFKA_TOPIC = 'aws.structured-data.articles.v1',
  KEY_FORMAT = 'AVRO',
  VALUE_FORMAT = 'AVRO'
);
