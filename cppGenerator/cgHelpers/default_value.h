#pragma once
#include <cstdint>
#include <string>

namespace cgHelpers {
    template<typename>
    struct default_value;

    template<>
    struct default_value<double> {
        static constexpr double value = 0;
    };

    template<>
    struct default_value<std::string> {
        static constexpr std::string value = "";
    };

    template<>
    struct default_value<uint8_t> {
        static constexpr uint8_t value = 0;
    };

    template<>
    struct default_value<uint16_t> {
        static constexpr uint16_t value = 0;
    };

    template<>
    struct default_value<uint32_t> {
        static constexpr uint32_t value = 0;
    };

    template<>
    struct default_value<uint64_t> {
        static constexpr uint64_t value = 0;
    };

    template<>
    struct default_value<int8_t> {
        static constexpr int8_t value = 0;
    };

    template<>
    struct default_value<int16_t> {
        static constexpr int16_t value = 0;
    };

    template<>
    struct default_value<int32_t> {
        static constexpr int32_t value = 0;
    };

    template<>
    struct default_value<int64_t> {
        static constexpr int64_t value = 0;
    };

    template<>
    struct default_value<bool> {
        static constexpr bool value = false;
    };
}
