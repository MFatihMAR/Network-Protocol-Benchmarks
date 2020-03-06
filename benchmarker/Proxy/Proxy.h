#ifndef PROXY_H
#define PROXY_H

#include <cstdint>
#include <optional>

struct ProxyConfig
{
	uint16_t ProxyPort;
	uint16_t NorthPort;
	uint16_t SouthPort;

	uint16_t SockBufSize;
};

enum class ProxyError
{
	CannotCreateSocket,
	CannotConfigureSocket,
	CannotBindSocket
};

class Proxy
{
public:
	bool IsRunning() const;
	std::optional<ProxyError> Listen(const ProxyConfig& config);
	std::pair<std::optional<ProxyError>, std::vector<uint8_t>> ReceiveNorth();
	std::optional<ProxyError> SendNorth(const std::vector<uint8_t>& data);
	std::pair<std::optional<ProxyError>, std::vector<uint8_t>> ReceiveSouth();
	std::optional<ProxyError> SendSouth(const std::vector<uint8_t>& data);
	void Stop();

private:
	int m_socket = -1;
};

#endif // PROXY_H
