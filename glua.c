
#include <luajit.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "_cgo_export.h"

extern int sync_go_method(lua_State* _L);
extern int async_go_method(lua_State* _L);

int gluaL_dostring(lua_State* _L, char* script) {
	return luaL_dostring(_L, script);
}
void glua_getglobal(lua_State* _L, char* name) {
	lua_getglobal(_L, name);
}
void glua_setglobal(lua_State* _L, char* name) {
	lua_setglobal(_L, name);
}
void glua_pushlightuserdata(lua_State* _L, void* obj) {
	lua_pushlightuserdata(_L, obj);
}
int glua_pcall(lua_State* _L, int args, int results) {
	return lua_pcall(_L, args, results, 0);
}
lua_Number glua_tonumber(lua_State* _L, int index) {
	return lua_tonumber(_L, index);
}
int glua_yield(lua_State *_L, int nresults) {
	return lua_yield(_L, nresults);
}
const char* glua_tostring(lua_State* _L, int index) {
	return lua_tostring(_L, index);
}
void glua_pop(lua_State* _L, int num) {
	lua_pop(_L, num);
}
lua_State *glua_tothread(lua_State* _L, int index) {
	return lua_tothread(_L, index);
}

int glua_istable(lua_State* _L, int index) {
	return lua_istable(_L, index);
}
void* glua_touserdata(lua_State* _L, int index) {
	return lua_touserdata(_L, index);
}

void register_go_method(lua_State* _L) {
	// lua_pushcfunction(_L, &sync_go_method);
	// lua_setglobal(_L, "sync_go_method");
	// lua_pushcfunction(_L, &async_go_method);
	// lua_setglobal(_L, "async_go_method");
}