package conversion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	playlistv0alpha1 "github.com/grafana/grafana/apps/playlist/pkg/apis/playlist/v0alpha1"
	playlistv1 "github.com/grafana/grafana/apps/playlist/pkg/apis/playlist/v1"
)

func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddConversionFunc((*playlistv0alpha1.Playlist)(nil), (*playlistv1.Playlist)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v0alpha1_Playlist_To_v1_Playlist(a.(*playlistv0alpha1.Playlist), b.(*playlistv1.Playlist), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*playlistv1.Playlist)(nil), (*playlistv0alpha1.Playlist)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_Playlist_To_v0alpha1_Playlist(a.(*playlistv1.Playlist), b.(*playlistv0alpha1.Playlist), scope)
	}); err != nil {
		return err
	}
	return nil
}

func Convert_v0alpha1_Playlist_To_v1_Playlist(in *playlistv0alpha1.Playlist, out *playlistv1.Playlist, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Spec.Title = in.Spec.Title
	out.Spec.Interval = in.Spec.Interval
	out.Spec.Items = make([]playlistv1.PlaylistItem, len(in.Spec.Items))
	for i, item := range in.Spec.Items {
		if item.Type == playlistv0alpha1.PlaylistItemTypeDashboardById {
			return fmt.Errorf("cannot convert dashboard by id to v1")
		}
		out.Spec.Items[i] = playlistv1.PlaylistItem{
			Type:  playlistv1.PlaylistItemType(item.Type),
			Value: item.Value,
		}
	}
	return nil
}

func Convert_v1_Playlist_To_v0alpha1_Playlist(in *playlistv1.Playlist, out *playlistv0alpha1.Playlist, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Spec.Title = in.Spec.Title
	out.Spec.Interval = in.Spec.Interval
	out.Spec.Items = make([]playlistv0alpha1.PlaylistItem, len(in.Spec.Items))
	for i, item := range in.Spec.Items {
		out.Spec.Items[i] = playlistv0alpha1.PlaylistItem{
			Type:  playlistv0alpha1.PlaylistItemType(item.Type),
			Value: item.Value,
		}
	}
	return nil
}
