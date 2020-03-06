#include <netinet/in.h>
#include <unistd.h>
#include <fcntl.h>
#include "Proxy.h"

std::optional<ProxyError> Proxy::Listen(const ProxyConfig& config)
{
	// TODO: check `config` arguments

	int proxySocket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
	if (proxySocket < 0)
	{
		return ProxyError::CannotCreateSocket;
	}

	int flags = fcntl(proxySocket, F_GETFL, 0);
	if (fcntl(proxySocket, F_SETFL, flags | O_NONBLOCK) < 0)
	{
		close(proxySocket);
		return ProxyError::CannotConfigureSocket;
	}

	sockaddr_in proxyAddress = {0};
	proxyAddress.sin_family = AF_INET;
	proxyAddress.sin_addr.s_addr = htonl(INADDR_ANY);
	proxyAddress.sin_port = htons(config.ProxyPort);
	if (bind(proxySocket, (sockaddr*) &proxyAddress, sizeof(proxyAddress)) < 0)
	{
		close(proxySocket);
		return ProxyError::CannotBindSocket;
	}

	m_socket = proxySocket;
	return std::optional<ProxyError>();
}
