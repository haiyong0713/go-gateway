package template

const (
	PomContent = `<?xml version="1.0" encoding="UTF-8"?>
<project xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd" xmlns="http://maven.apache.org/POM/4.0.0"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <modelVersion>4.0.0</modelVersion>
  <groupId>{{.AppKey}}</groupId>
  <artifactId>{{.BundleName}}</artifactId>
  <version>{{.CIJobID}}</version>
  <packaging>bbr</packaging>
  <name>{{.BundleName}}</name>
  <description>Bili Bundle Archive</description>
</project>`

	BundleName    = `{{.BundleName}}-{{.CIJobID}}.bbr`
	BundleMd5     = `{{.BundleName}}-{{.CIJobID}}.bbr.md5`
	BundleSha1    = `{{.BundleName}}-{{.CIJobID}}.bbr.sha1`
	BundlePom     = `{{.BundleName}}-{{.CIJobID}}.pom`
	BundlePomMd5  = `{{.BundleName}}-{{.CIJobID}}.pom.md5`
	BundlePomSha1 = `{{.BundleName}}-{{.CIJobID}}.pom.sha1`
)

type PomData struct {
	AppKey     string
	BundleName string
	CIJobID    int64
}
