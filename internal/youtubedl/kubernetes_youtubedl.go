package youtubedl

import (
	"github.com/mitaka8/playlist-bot/internal/randstr"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"go.uber.org/fx"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesYouTubeDL struct {
	Configuration Configuration
	Kubernetes    *kubernetes.KubernetesClient
	Log           *zap.Logger
}

type NewKubernetesYouTubeDLParams struct {
	fx.In

	Configuration Configuration
	Kubernetes    *kubernetes.KubernetesClient
	Log           *zap.Logger
}

func NewKubernetesYouTubeDL(p NewKubernetesYouTubeDLParams) *KubernetesYouTubeDL {
	return &KubernetesYouTubeDL{
		Configuration: p.Configuration,
		Kubernetes:    p.Kubernetes,
		Log:           p.Log,
	}
}

func (e *KubernetesYouTubeDL) Save(source string, guildId string, uuid string) {
	err := e.Kubernetes.RunJob(convertJob(source, guildId, uuid))
	if err != nil {
		e.Log.Error("Error running job on Kubernetes", zap.Error(err))
	}

}

func convertJob(url, guildId, uuid string) *batchv1.Job {

	id := randstr.Concat(
		randstr.Randstr(randstr.Lowercase, 1),
		randstr.Randstr(randstr.Lowercase+randstr.Numbers, 5),
	)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "playlist-bot-download-" + id,
			Namespace: "playlist-bot",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "downloader",
							Image: "mitaka8/playlist-bot:latest",
							Command: []string{
								"playlist-bot",
								"save",
								url,
								guildId,
								uuid,
							},
							Env: []v1.EnvVar{
								{
									Name:  "YOUTUBE_IMPLEMENTATION",
									Value: "exec",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/discordbot",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "playlist-bot",
									},
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}
}
