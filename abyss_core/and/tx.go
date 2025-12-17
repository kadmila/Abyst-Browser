package and

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/kadmila/Abyss-Browser/abyss_core/ahmp"
	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
	"github.com/kadmila/Abyss-Browser/abyss_core/tools/functional"
)

func SendJN(peer ani.IAbyssPeer, local_session_id uuid.UUID, path string, timestamp time.Time) error {
	return peer.Send(ahmp.JN_T, RawJN{
		SenderSessionID: local_session_id.String(),
		Text:            path,
		TimeStamp:       timestamp.UnixMilli(),
	})
}
func SendJOK(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, timestamp time.Time, world_url string, member_sessions []PeerWorldSession) error {
	return peer.Send(ahmp.JOK_T, RawJOK{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		TimeStamp:       timestamp.UnixMilli(),
		Text:            world_url,
		Neighbors: functional.Filter(member_sessions, func(session PeerWorldSession) RawSessionInfoForDiscovery {
			return RawSessionInfoForDiscovery{
				PeerID:                     session.Peer.ID(),
				AddressCandidates:          functional.Filter(session.AddressCandidates, func(a netip.AddrPort) string { return a.String() }),
				SessionID:                  session.SessionID.String(),
				TimeStamp:                  session.TimeStamp.UnixMilli(),
				RootCertificateDer:         session.Peer.RootCertificateDer(),
				HandshakeKeyCertificateDer: session.Peer.HandshakeKeyCertificateDer(),
			}
		}),
	})
}
func SendJDN(peer ani.IAbyssPeer, peer_session_id uuid.UUID, code int, message string) error {
	return peer.Send(ahmp.JDN_T, RawJDN{
		RecverSessionID: peer_session_id.String(),
		Text:            message,
		Code:            code,
	})
}
func SendJNI(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, member_session PeerWorldSession) error {
	return peer.Send(ahmp.JNI_T, RawJNI{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		Neighbor: RawSessionInfoForDiscovery{
			PeerID:                     member_session.Peer.ID(),
			AddressCandidates:          functional.Filter(member_session.AddressCandidates, func(a netip.AddrPort) string { return a.String() }),
			SessionID:                  member_session.SessionID.String(),
			TimeStamp:                  member_session.TimeStamp.UnixMilli(),
			RootCertificateDer:         member_session.Peer.RootCertificateDer(),
			HandshakeKeyCertificateDer: member_session.Peer.HandshakeKeyCertificateDer(),
		},
	})
}
func SendMEM(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, timestamp time.Time) error {
	return peer.Send(ahmp.MEM_T, RawMEM{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		TimeStamp:       timestamp.UnixMilli(),
	})
}
func SendSJN(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, member_sessions []ANDPeerSessionIdentity) error {
	return peer.Send(ahmp.SJN_T, RawSJN{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		MemberInfos: functional.Filter(member_sessions, func(i ANDPeerSessionIdentity) RawSessionInfoForSJN {
			return RawSessionInfoForSJN{
				PeerID:    i.PeerID,
				SessionID: i.SessionID.String(),
			}
		}),
	})
}
func SendCRR(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, member_sessions []ANDPeerSessionIdentity) error {
	return peer.Send(ahmp.CRR_T, RawCRR{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		MemberInfos: functional.Filter(member_sessions, func(i ANDPeerSessionIdentity) RawSessionInfoForSJN {
			return RawSessionInfoForSJN{
				PeerID:    i.PeerID,
				SessionID: i.SessionID.String(),
			}
		}),
	})
}
func SendRST(peer ani.IAbyssPeer, local_session_id uuid.UUID, peer_session_id uuid.UUID, message string) error {
	return peer.Send(ahmp.RST_T, RawRST{
		SenderSessionID: local_session_id.String(),
		RecverSessionID: peer_session_id.String(),
		Message:         message,
	})
}
