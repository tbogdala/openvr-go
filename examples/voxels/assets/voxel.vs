#version 330
precision highp float;

uniform mat4 MVP_MATRIX;
uniform mat4 M_MATRIX;
in vec3 VERTEX_POSITION;
in vec3 VERTEX_NORMAL;
in vec2 VERTEX_UV_0;
in float VERTEX_TEXTURE_INDEX;
in float VERTEX_VOXEL_BF;

flat out int vs_vert_texindex;
out float vs_vert_ao;
out float vs_vert_ao_corner;
out vec2 vs_uvcoord;
out vec4 w_position;
out vec3 w_normal;

void main()
{
    vs_uvcoord = VERTEX_UV_0;

    int bitfield = int(VERTEX_VOXEL_BF);
    vs_vert_texindex = int(VERTEX_TEXTURE_INDEX);
    vs_vert_ao = float(bitfield & 0x01);
    vs_vert_ao_corner = float((bitfield & 0x02) >> 1);

    vec4 vert4 = vec4(VERTEX_POSITION, 1.0);
    mat3 normal_mat = transpose(inverse(mat3(M_MATRIX)));
    w_position = M_MATRIX * vert4;
    w_normal = normal_mat * VERTEX_NORMAL;

    gl_Position = MVP_MATRIX * vert4;
}
