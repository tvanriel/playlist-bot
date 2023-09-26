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

func (e *KubernetesYouTubeDL) Save(p YouTubeDLParams) error {
	err := e.Kubernetes.RunJob(convertJob(p.Source, p.GuildID, p.PlaylistName))
	if err != nil {
		e.Log.Error("Error running job on Kubernetes", zap.Error(err))
	}
	return err

}

func convertJob(url, guildId, playlistName string) *batchv1.Job {

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
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
                                                "app.kubernetes.io/part-of": "playlist-bot",
                                                "app.kubernetes.io/name": "playlist-bot",
                                                "app.kubernetes.io/version": "latest",
                                                "app.kubernetes.io/component": "downloader",
                                                "app.kubernetes.io/instance": "downloader-" + id,
						"app": "playlist-bot-downloader",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "downloader",
							Image: "mitaka8/playlist-bot:latest",
							Command: []string{
								"playlist-bot",
								"save",
								"-u", url,
								"-g", guildId,
								"-p", playlistName,
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
					// Make sure only one can run at a time.  To avoid throttling the CPU too far.
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "playlist-bot-downloader",
										},
									},
									TopologyKey: "kubernetes.io/hostname",
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
