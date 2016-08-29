#version 330
precision highp float;

uniform vec4 MATERIAL_DIFFUSE;
uniform vec4 MATERIAL_SPECULAR;
uniform float MATERIAL_SHININESS;

uniform vec3 LIGHT_POSITION[4];
uniform vec4 LIGHT_DIFFUSE[4];
uniform float LIGHT_DIFFUSE_INTENSITY[4];
uniform float LIGHT_AMBIENT_INTENSITY[4];
uniform float LIGHT_SPECULAR_INTENSITY[4];
uniform vec3 LIGHT_DIRECTION[4];
uniform int LIGHT_COUNT;
uniform vec3 CAMERA_WORLD_POSITION;

uniform sampler2DArray VOXEL_TEXTURES;

flat in int vs_vert_texindex;
in float vs_vert_ao;
in float vs_vert_ao_corner;

in vec2 vs_uvcoord;
in vec4 w_position;
in vec3 w_normal;

out vec4 frag_color;

const float Epsilon = 0.0001;
const int MAX_LIGHT = 4;

// Gamma correction routines
const float gamma = 2.2;

vec3 toLinear(vec3 c) {
    return pow(c, vec3(gamma));
}
vec4 toLinear(vec4 c) {
    return pow(c, vec4(gamma));
}

vec3 toGamma(vec3 c) {
    return pow(c, vec3(1.0 / gamma));
}
vec4 toGamma(vec4 c) {
    return pow(c, vec4(1.0 / gamma));
}

// TODO: fix all the variables captured from global namespace
vec4 PhongShading(int light_i)
{
    vec3 l_light_color = LIGHT_DIFFUSE[light_i].rgb;
    vec3 l_specular_color = MATERIAL_SPECULAR.xyz;
    vec4 texColor = texture(VOXEL_TEXTURES, vec3(vs_uvcoord, vs_vert_texindex));
    vec3 l_base_color = toLinear(texColor.rgb);

    // world space normal
    vec3 N = normalize(w_normal.xyz);

    // calculate the direction towards the light in world space
    vec3 L;
    // if light direction is not set, calculate it from the position
    if (abs(LIGHT_DIRECTION[light_i].x) < Epsilon && abs(LIGHT_DIRECTION[light_i].y) < Epsilon && abs(LIGHT_DIRECTION[light_i].z) < Epsilon) {
      L = normalize(LIGHT_POSITION[light_i] - w_position.xyz);
    } else {
      // otherwise we just use the direction here
      L = normalize(-LIGHT_DIRECTION[light_i]);
    }

    // get the diffuse intensity
    float Id = max(0.0, dot(N, L));

    // calculate the specular coefficient appropriate based on diffuse intensity
    float S = 0.0;
    if (Id > 0.0 && MATERIAL_SHININESS > 0.0) {
        // calculate the specular
        vec3 R = normalize(reflect(-L, N));

        // calculate surface to camera in world space
        vec3 E = normalize(CAMERA_WORLD_POSITION - w_position.xyz);

        S = pow(max(0.0, dot(E, R)), MATERIAL_SHININESS);
    }

    vec3 final_ambient = LIGHT_AMBIENT_INTENSITY[light_i] * l_base_color * l_light_color;
    vec3 final_diffuse = LIGHT_DIFFUSE_INTENSITY[light_i] * l_base_color * l_light_color * Id;
    vec3 final_specular = LIGHT_SPECULAR_INTENSITY[light_i] * l_specular_color.xyz * l_light_color * S;
    clamp(final_specular, 0.0, 1.0);

    return vec4(final_ambient + final_diffuse + final_specular, MATERIAL_DIFFUSE.a);
}


void main()
{
  vec4 final_color;


  for (int i=0; i<MAX_LIGHT; ++i) {
    if (i >= LIGHT_COUNT) {
      break;
    }

    final_color += PhongShading(i);
  }

  // darken based on ao
  float aoFactorA = max(0.0, vs_vert_ao-0.4);
  float aoFactorB = max(0.0, vs_vert_ao_corner-0.4);
  float aoFactor = max(aoFactorA, aoFactorB);
  final_color = final_color * (1.0 - aoFactor);

  frag_color =  toGamma(final_color);
  frag_color.a = 1.0;

  //frag_color = vec4(1.0, 1.0, 1.0, 1.0);
}
